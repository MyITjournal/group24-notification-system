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
} from '@nestjs/common';
import { SimpleUsersService } from './simple-users.service';
import {
  CreateSimpleUserInput,
  SimpleUserResponse,
  SimpleUserPreferencesResponse,
  BatchGetSimpleUserPreferencesInput,
  BatchGetSimpleUserPreferencesResponse,
  UpdateLastNotificationInput,
  UpdateSimpleUserPreferencesInput,
} from './dto/simple-user.dto';

@Controller('api/v1/users')
export class SimpleUsersController {
  constructor(private readonly simpleUsersService: SimpleUsersService) {}

  @Get()
  async getAllUsers(): Promise<SimpleUserPreferencesResponse[]> {
    try {
      return await this.simpleUsersService.getAllUsers();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      throw new HttpException(
        {
          error: {
            code: 'USERS_FETCH_FAILED',
            message: 'Failed to fetch users',
            details: {
              error: errorMessage,
            },
          },
        },
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Post()
  async createUser(
    @Body() input: CreateSimpleUserInput,
  ): Promise<SimpleUserResponse> {
    console.log('SimpleUsersController.createUser received:', input);
    try {
      return await this.simpleUsersService.createUser(input);
    } catch (error) {
      const err = error as { code?: string; message?: string };
      if (err.code === '23505') {
        // Unique constraint violation (email already exists)
        throw new HttpException(
          {
            error: {
              code: 'EMAIL_ALREADY_EXISTS',
              message: 'A user with this email already exists',
              details: {
                email: input.email,
              },
            },
          },
          HttpStatus.CONFLICT,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        {
          error: {
            code: 'USER_CREATION_FAILED',
            message: 'Failed to create user',
            details: {
              error: errorMessage,
            },
          },
        },
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Get(':user_id/preferences')
  async getUserPreferences(
    @Param('user_id') userId: string,
  ): Promise<SimpleUserPreferencesResponse> {
    try {
      return await this.simpleUsersService.getUserPreferences(userId);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { status?: number; message?: string };
      if (err.status === 404 || err.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          {
            error: {
              code: 'USER_NOT_FOUND',
              message: `User with ID ${userId} does not exist`,
              details: {
                user_id: userId,
              },
            },
          },
          HttpStatus.NOT_FOUND,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        {
          error: {
            code: 'PREFERENCES_FETCH_FAILED',
            message: 'Failed to fetch user preferences',
            details: {
              error: errorMessage,
            },
          },
        },
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  @Post('preferences/batch')
  async batchGetUserPreferences(
    @Body() input: BatchGetSimpleUserPreferencesInput,
  ): Promise<BatchGetSimpleUserPreferencesResponse> {
    try {
      // Check for duplicates
      const uniqueIds = new Set(input.user_ids);
      if (uniqueIds.size !== input.user_ids.length) {
        throw new HttpException(
          {
            error: {
              code: 'DUPLICATE_USER_IDS',
              message: 'The user_ids array contains duplicate values',
              details: {
                total_provided: input.user_ids.length,
                unique_count: uniqueIds.size,
              },
            },
          },
          HttpStatus.BAD_REQUEST,
        );
      }

      return await this.simpleUsersService.batchGetUserPreferences(input);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { message?: string };
      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        {
          error: {
            code: 'BATCH_FETCH_FAILED',
            message: 'Failed to fetch batch user preferences',
            details: {
              error: errorMessage,
            },
          },
        },
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
  ): Promise<SimpleUserPreferencesResponse> {
    try {
      // Validate that at least one field is provided
      if (input.email === undefined && input.push === undefined) {
        throw new HttpException(
          {
            error: {
              code: 'NO_FIELDS_PROVIDED',
              message:
                'At least one preference field (email or push) must be provided',
              details: {
                provided_fields: Object.keys(input),
              },
            },
          },
          HttpStatus.BAD_REQUEST,
        );
      }

      return await this.simpleUsersService.updateUserPreferences(userId, input);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      const err = error as { status?: number; message?: string };
      if (err.status === 404 || err.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          {
            error: {
              code: 'USER_NOT_FOUND',
              message: `User with ID ${userId} does not exist`,
              details: {
                user_id: userId,
              },
            },
          },
          HttpStatus.NOT_FOUND,
        );
      }

      const errorMessage = err.message ?? 'Unknown error';
      throw new HttpException(
        {
          error: {
            code: 'PREFERENCES_UPDATE_FAILED',
            message: 'Failed to update user preferences',
            details: {
              error: errorMessage,
            },
          },
        },
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }
}
