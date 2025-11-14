import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  HttpException,
  HttpStatus,
  HttpCode,
  Patch,
  Query,
  ParseIntPipe,
  DefaultValuePipe,
  Headers,
  UseGuards,
  SetMetadata,
} from '@nestjs/common';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';
import { SimpleUsersService } from './simple_users.service';
import { CacheService } from '../cache/cache_service';
import {
  CreateSimpleUserInput,
  SimpleUserResponse,
  SimpleUserPreferencesResponse,
  BatchGetSimpleUserPreferencesInput,
  BatchGetSimpleUserPreferencesResponse,
  UpdateLastNotificationInput,
  UpdateSimpleUserPreferencesInput,
  ApiResponse,
} from './dto/simple_user.dto';

export const Public = () => SetMetadata('isPublic', true);

@Controller('api/v1/users')
@UseGuards(JwtAuthGuard)
export class SimpleUsersController {
  constructor(
    private readonly simpleUsersService: SimpleUsersService,
    private readonly cacheService: CacheService,
  ) {}

  @Get()
  async getAllUsers(
    @Query('page', new DefaultValuePipe(1), ParseIntPipe) page: number,
    @Query('limit', new DefaultValuePipe(10), ParseIntPipe) limit: number,
  ): Promise<ApiResponse<SimpleUserPreferencesResponse[]>> {
    try {
      // Validate pagination parameters
      if (page < 1) {
        throw new HttpException(
          ApiResponse.error('Page must be greater than 0', 'INVALID_PAGE'),
          HttpStatus.BAD_REQUEST,
        );
      }
      if (limit < 1 || limit > 100) {
        throw new HttpException(
          ApiResponse.error('Limit must be between 1 and 100', 'INVALID_LIMIT'),
          HttpStatus.BAD_REQUEST,
        );
      }

      const { users, total } = await this.simpleUsersService.getAllUsers(
        page,
        limit,
      );

      const total_pages = Math.ceil(total / limit);
      const meta = {
        total,
        limit,
        page,
        total_pages,
        has_next: page < total_pages,
        has_previous: page > 1,
      };

      return ApiResponse.success('Users retrieved successfully', users, meta);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      throw new HttpException(
        ApiResponse.error('Failed to fetch users', errorMessage),
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Post()
  @Public() // Allow user registration without authentication
  async createUser(
    @Body() input: CreateSimpleUserInput,
    @Headers('x-request-id') requestId?: string,
  ): Promise<ApiResponse<SimpleUserResponse>> {
    console.log('SimpleUsersController.createUser received:', input);

    // Check for idempotency if request ID is provided
    if (requestId) {
      const cachedResponse =
        await this.cacheService.getIdempotentResponse(requestId);
      if (cachedResponse) {
        console.log(`Idempotent request detected: ${requestId}`);
        return cachedResponse as ApiResponse<SimpleUserResponse>;
      }
    }

    try {
      const user = await this.simpleUsersService.createUser(input);
      const response = ApiResponse.success('User created successfully', user);

      // Cache the response for idempotency (24 hours TTL)
      if (requestId) {
        await this.cacheService.setIdempotentResponse(
          requestId,
          response,
          86400,
        );
      }

      return response;
    } catch (error) {
      const err = error as { code?: string; message?: string };
      if (err.code === '23505') {
        // Unique constraint violation (email already exists)
        throw new HttpException(
          ApiResponse.error(
            'A user with this email already exists',
            'EMAIL_ALREADY_EXISTS',
          ),
          HttpStatus.CONFLICT,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        ApiResponse.error('Failed to create user', errorMessage),
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Get(':user_id/preferences')
  async getUserPreferences(
    @Param('user_id') userId: string,
  ): Promise<ApiResponse<SimpleUserPreferencesResponse>> {
    try {
      const preferences =
        await this.simpleUsersService.getUserPreferences(userId);
      return ApiResponse.success(
        'User preferences retrieved successfully',
        preferences,
      );
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { status?: number; message?: string };
      if (err.status === 404 || err.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          ApiResponse.error(
            `User with ID ${userId} does not exist`,
            'USER_NOT_FOUND',
          ),
          HttpStatus.NOT_FOUND,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        ApiResponse.error('Failed to fetch user preferences', errorMessage),
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Post('preferences/batch')
  async batchGetUserPreferences(
    @Body() input: BatchGetSimpleUserPreferencesInput,
    @Headers('x-request-id') requestId?: string,
  ): Promise<ApiResponse<BatchGetSimpleUserPreferencesResponse>> {
    // Check for idempotency if request ID is provided
    if (requestId) {
      const cachedResponse =
        await this.cacheService.getIdempotentResponse(requestId);
      if (cachedResponse) {
        console.log(`Idempotent request detected: ${requestId}`);
        return cachedResponse as ApiResponse<BatchGetSimpleUserPreferencesResponse>;
      }
    }

    try {
      // Check for duplicates
      const uniqueIds = new Set(input.user_ids);
      if (uniqueIds.size !== input.user_ids.length) {
        throw new HttpException(
          ApiResponse.error(
            'The user_ids array contains duplicate values',
            'DUPLICATE_USER_IDS',
          ),
          HttpStatus.BAD_REQUEST,
        );
      }

      const result =
        await this.simpleUsersService.batchGetUserPreferences(input);
      const response = ApiResponse.success(
        'Batch user preferences retrieved successfully',
        result,
      );

      // Cache the response for idempotency (24 hours TTL)
      if (requestId) {
        await this.cacheService.setIdempotentResponse(
          requestId,
          response,
          86400,
        );
      }

      return response;
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { message?: string };
      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        ApiResponse.error(
          'Failed to fetch batch user preferences',
          errorMessage,
        ),
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Post(':user_id/last-notification')
  @HttpCode(204)
  updateLastNotification(
    @Param('user_id') userId: string,
    @Body() input: UpdateLastNotificationInput,
  ): void {
    // Fire-and-forget - don't wait for completion
    this.simpleUsersService
      .updateLastNotificationTime(userId, input)
      .catch((error) => {
        console.error(
          `Fire-and-forget notification update failed for user ${userId}:`,
          error,
        );
      });

    // Return immediately
    return;
  }

  @Patch(':user_id/preferences')
  async updateUserPreferences(
    @Param('user_id') userId: string,
    @Body() input: UpdateSimpleUserPreferencesInput,
  ): Promise<ApiResponse<SimpleUserPreferencesResponse>> {
    try {
      // Validate that at least one field is provided
      if (input.email === undefined && input.push === undefined) {
        throw new HttpException(
          ApiResponse.error(
            'At least one preference field (email or push) must be provided',
            'NO_FIELDS_PROVIDED',
          ),
          HttpStatus.BAD_REQUEST,
        );
      }

      const result = await this.simpleUsersService.updateUserPreferences(
        userId,
        input,
      );
      return ApiResponse.success(
        'User preferences updated successfully',
        result,
      );
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { status?: number; message?: string };
      if (err.status === 404 || err.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          ApiResponse.error(
            `User with ID ${userId} does not exist`,
            'USER_NOT_FOUND',
          ),
          HttpStatus.NOT_FOUND,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        ApiResponse.error('Failed to update user preferences', errorMessage),
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }
}
