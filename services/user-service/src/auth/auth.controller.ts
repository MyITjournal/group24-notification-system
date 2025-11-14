import { Controller, Post, Body, HttpCode, HttpStatus } from '@nestjs/common';
import type { LoginDto, LoginResponse } from './auth.service';
import { AuthService } from './auth.service';
import { ApiResponse } from '../simple_users/dto/simple_user.dto';

@Controller('api/v1/auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('login')
  @HttpCode(HttpStatus.OK)
  async login(@Body() loginDto: LoginDto): Promise<ApiResponse<LoginResponse>> {
    const result = await this.authService.login(loginDto);
    return ApiResponse.success('Login successful', result);
  }
}
