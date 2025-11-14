import { Injectable, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { SimpleUsersService } from '../simple_users/simple_users.service';
import * as bcrypt from 'bcrypt';

export interface LoginDto {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: {
    user_id: string;
    name: string;
    email: string;
  };
}

export interface JwtPayload {
  sub: string; // user_id
  email: string;
  name: string;
  iat?: number;
  exp?: number;
}

@Injectable()
export class AuthService {
  constructor(
    private readonly simpleUsersService: SimpleUsersService,
    private readonly jwtService: JwtService,
  ) {}

  async validateUser(email: string, password: string): Promise<any> {
    try {
      const user = await this.simpleUsersService.getUserByEmail(email);

      if (!user) {
        return null;
      }

      const isPasswordValid = await bcrypt.compare(password, user.password);

      if (!isPasswordValid) {
        return null;
      }

      // Return user without password
      const { password: _, ...result } = user;
      return result;
    } catch (error) {
      return null;
    }
  }

  async login(loginDto: LoginDto): Promise<LoginResponse> {
    const user = await this.validateUser(loginDto.email, loginDto.password);

    if (!user) {
      throw new UnauthorizedException({
        success: false,
        error: 'INVALID_CREDENTIALS',
        message: 'Invalid email or password',
        meta: null,
      });
    }

    const payload: JwtPayload = {
      sub: user.user_id,
      email: user.email,
      name: user.name,
    };

    const access_token = this.jwtService.sign(payload);

    return {
      access_token,
      token_type: 'Bearer',
      expires_in: 86400, // 24 hours in seconds
      user: {
        user_id: user.user_id,
        name: user.name,
        email: user.email,
      },
    };
  }

  async verifyToken(token: string): Promise<JwtPayload> {
    try {
      return this.jwtService.verify<JwtPayload>(token);
    } catch (error) {
      throw new UnauthorizedException({
        success: false,
        error: 'INVALID_TOKEN',
        message: 'Invalid or expired token',
        meta: null,
      });
    }
  }
}
