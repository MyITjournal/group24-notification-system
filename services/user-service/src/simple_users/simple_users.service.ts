import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, In } from 'typeorm';
import { SimpleUser } from './entity/simple_user.entity';
import {
  CreateSimpleUserInput,
  SimpleUserResponse,
  SimpleUserPreferencesResponse,
  BatchGetSimpleUserPreferencesInput,
  BatchGetSimpleUserPreferencesResponse,
  UpdateLastNotificationInput,
  UpdateSimpleUserPreferencesInput,
} from './dto/simple_user.dto';
import * as bcrypt from 'bcrypt';
import { CacheService } from '../cache/cache_service';

@Injectable()
export class SimpleUsersService {
  constructor(
    @InjectRepository(SimpleUser)
    private readonly simpleUserRepository: Repository<SimpleUser>,
    private readonly cacheService: CacheService,
  ) {}

  async getUserByEmail(email: string): Promise<SimpleUser | null> {
    return await this.simpleUserRepository.findOne({
      where: { email },
    });
  }

  async createUser(input: CreateSimpleUserInput): Promise<SimpleUserResponse> {
    console.log('SimpleUsersService.createUser called with:', input);

    // Generate user ID
    const userId = `usr_${Math.random().toString(36).substring(2, 10)}`;

    // Hash password
    const saltRounds = 10;
    const hashedPassword = await bcrypt.hash(input.password, saltRounds);

    // Create simple user
    const user = this.simpleUserRepository.create({
      user_id: userId,
      name: input.name,
      email: input.email,
      password: hashedPassword,
      push_token: input.push_token,
      email_preference: input.preferences.email,
      push_preference: input.preferences.push,
    });

    await this.simpleUserRepository.save(user);

    // Return response (without password)
    return {
      user_id: user.user_id,
      name: user.name,
      email: user.email,
      push_token: user.push_token,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      created_at: user.created_at,
    };
  }

  async getUserPreferences(
    userId: string,
  ): Promise<SimpleUserPreferencesResponse> {
    // Try to get from cache first
    const cached = (await this.cacheService.getUserPreferences(
      userId,
    )) as SimpleUserPreferencesResponse | null;
    if (cached) {
      console.log(`Cache HIT for user preferences: ${userId}`);
      return cached;
    }

    console.log(`Cache MISS for user preferences: ${userId}`);

    const user = await this.simpleUserRepository.findOne({
      where: { user_id: userId },
    });

    if (!user) {
      throw new NotFoundException({
        code: 'USER_NOT_FOUND',
        message: `User with ID ${userId} does not exist`,
        details: {
          user_id: userId,
        },
      });
    }

    const response = {
      user_id: user.user_id,
      email: user.email,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      last_notification_email: user.last_notification_email,
      last_notification_push: user.last_notification_push,
      last_notification_id: user.last_notification_id,
      updated_at: user.updated_at,
    };

    // Cache the result (1 hour TTL)
    await this.cacheService.setUserPreferences(userId, response, 3600);

    return response;
  }

  async batchGetUserPreferences(
    input: BatchGetSimpleUserPreferencesInput,
  ): Promise<BatchGetSimpleUserPreferencesResponse> {
    // Remove duplicates from user_ids
    const uniqueUserIds = [...new Set(input.user_ids)];

    // Try to get from cache first
    const cachedPreferences =
      await this.cacheService.getBatchUserPreferences(uniqueUserIds);

    console.log(
      `Cache HIT for ${cachedPreferences.size} out of ${uniqueUserIds.length} users`,
    );

    // Find user IDs not in cache
    const uncachedUserIds = uniqueUserIds.filter(
      (id) => !cachedPreferences.has(id),
    );

    // Fetch uncached users from database
    const users =
      uncachedUserIds.length > 0
        ? await this.simpleUserRepository.find({
            where: { user_id: In(uncachedUserIds) },
          })
        : [];

    console.log(
      `Fetched ${users.length} users from database for ${uncachedUserIds.length} uncached IDs`,
    );

    // Build response for database users
    const dbUsersMap = new Map<string, SimpleUserPreferencesResponse>();
    users.forEach((user) => {
      const response = {
        user_id: user.user_id,
        email: user.email,
        preferences: {
          email: user.email_preference,
          push: user.push_preference,
        },
        last_notification_email: user.last_notification_email,
        last_notification_push: user.last_notification_push,
        last_notification_id: user.last_notification_id,
        updated_at: user.updated_at,
      };
      dbUsersMap.set(user.user_id, response);
    });

    // Cache the database results (1 hour TTL)
    if (dbUsersMap.size > 0) {
      await this.cacheService.setBatchUserPreferences(dbUsersMap, 3600);
    }

    // Combine cached and database results
    const cachedUsers = Array.from(
      cachedPreferences.values(),
    ) as SimpleUserPreferencesResponse[];
    const dbUsers = Array.from(dbUsersMap.values());
    const allUsers = [...cachedUsers, ...dbUsers];

    // Find not found user IDs
    const foundUserIds = new Set(allUsers.map((u) => u.user_id));
    const notFound = uniqueUserIds.filter((id) => !foundUserIds.has(id));

    return {
      users: allUsers,
      not_found: notFound,
      total_requested: uniqueUserIds.length,
      total_found: allUsers.length,
    };
  }

  async updateLastNotificationTime(
    userId: string,
    input: UpdateLastNotificationInput,
  ): Promise<void> {
    const user = await this.simpleUserRepository.findOne({
      where: { user_id: userId },
    });

    if (!user) {
      // Silently ignore if user doesn't exist (fire-and-forget)
      return;
    }

    // Update the appropriate last notification field
    if (input.channel === 'email') {
      user.last_notification_email = new Date(input.sent_at);
    } else if (input.channel === 'push') {
      user.last_notification_push = new Date(input.sent_at);
    }

    user.last_notification_id = input.notification_id;

    // Save and invalidate cache (fire-and-forget optimization)
    this.simpleUserRepository
      .save(user)
      .then(() => {
        // Invalidate cache after successful save
        return this.cacheService.invalidateUserPreferences(userId);
      })
      .catch((error) => {
        // Log error but don't throw (fire-and-forget)
        console.error(
          `Failed to update last notification time for user ${userId}:`,
          error,
        );
      });
  }

  async getAllUsers(
    page: number,
    limit: number,
  ): Promise<{ users: SimpleUserPreferencesResponse[]; total: number }> {
    // Use findAndCount for pagination
    const [users, total] = await this.simpleUserRepository.findAndCount({
      order: {
        created_at: 'DESC',
      },
      skip: (page - 1) * limit,
      take: limit,
    });

    const mappedUsers = users.map((user) => ({
      user_id: user.user_id,
      email: user.email,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      last_notification_email: user.last_notification_email,
      last_notification_push: user.last_notification_push,
      last_notification_id: user.last_notification_id,
      updated_at: user.updated_at,
    }));

    return {
      users: mappedUsers,
      total,
    };
  }

  async updateUserPreferences(
    userId: string,
    input: UpdateSimpleUserPreferencesInput,
  ): Promise<SimpleUserPreferencesResponse> {
    const user = await this.simpleUserRepository.findOne({
      where: { user_id: userId },
    });

    if (!user) {
      throw new NotFoundException({
        code: 'USER_NOT_FOUND',
        message: `User with ID ${userId} does not exist`,
        details: {
          user_id: userId,
        },
      });
    }

    // Update only provided fields
    if (input.email !== undefined) {
      user.email_preference = input.email;
    }

    if (input.push !== undefined) {
      user.push_preference = input.push;
    }

    // Save updated user
    await this.simpleUserRepository.save(user);

    // Invalidate cache after successful update
    await this.cacheService.invalidateUserPreferences(userId);

    // Return updated preferences
    return {
      user_id: user.user_id,
      email: user.email,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      last_notification_email: user.last_notification_email,
      last_notification_push: user.last_notification_push,
      last_notification_id: user.last_notification_id,
      updated_at: user.updated_at,
    };
  }
}
