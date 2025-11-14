import {
  IsString,
  IsEmail,
  IsBoolean,
  IsOptional,
  ValidateNested,
  IsArray,
  ArrayMaxSize,
  IsNotEmpty,
  IsDateString,
  IsIn,
} from 'class-validator';
import { Type } from 'class-transformer';

// ============ Preferences DTOs ============

export class SimpleUserPreferenceDto {
  @IsBoolean()
  email: boolean;

  @IsBoolean()
  push: boolean;
}

// ============ Create User DTOs ============

export class CreateSimpleUserInput {
  @IsString()
  @IsNotEmpty()
  name: string;

  @IsEmail()
  email: string;

  @IsString()
  @IsNotEmpty()
  password: string;

  @IsOptional()
  @IsString()
  push_token?: string;

  @ValidateNested()
  @Type(() => SimpleUserPreferenceDto)
  preferences: SimpleUserPreferenceDto;
}

export class SimpleUserResponse {
  user_id: string;
  name: string;
  email: string;
  push_token?: string;
  preferences: {
    email: boolean;
    push: boolean;
  };
  created_at: Date;
}

// ============ Get Preferences DTOs ============

export class SimpleUserPreferencesResponse {
  user_id: string;
  email: string;
  preferences: {
    email: boolean;
    push: boolean;
  };
  last_notification_email?: Date;
  last_notification_push?: Date;
  last_notification_id?: string;
  updated_at: Date;
}

// ============ Batch Get Preferences DTOs ============

export class BatchGetSimpleUserPreferencesInput {
  @IsArray()
  @ArrayMaxSize(100)
  @IsString({ each: true })
  user_ids: string[];
}

export class BatchGetSimpleUserPreferencesResponse {
  users: SimpleUserPreferencesResponse[];
  not_found: string[];
  total_requested: number;
  total_found: number;
}

// ============ Update Preferences DTOs ============

export class UpdateSimpleUserPreferencesInput {
  @IsOptional()
  @IsBoolean()
  email?: boolean;

  @IsOptional()
  @IsBoolean()
  push?: boolean;
}

// ============ Update Last Notification DTOs ============

export class UpdateLastNotificationInput {
  @IsIn(['email', 'push'])
  channel: 'email' | 'push';

  @IsDateString()
  sent_at: string;

  @IsString()
  notification_id: string;
}
