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

  async validateUser(
    email: string,
    password: string,
  ): Promise<Omit<any, 'password'> | null> {
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
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { password: _password, ...result } = user;
      return result;
    } catch {
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

    // Type guard: validateUser returns user object with required fields
    const validatedUser = user as {
      user_id: string;
      email: string;
      name: string;
    };

    const payload: JwtPayload = {
      sub: validatedUser.user_id,
      email: validatedUser.email,
      name: validatedUser.name,
    };

    const access_token = this.jwtService.sign(payload);

    return {
      access_token,
      token_type: 'Bearer',
      expires_in: 86400, // 24 hours in seconds
      user: {
        user_id: validatedUser.user_id,
        name: validatedUser.name,
        email: validatedUser.email,
      },
    };
  }

  verifyToken(token: string): JwtPayload {
    try {
      return this.jwtService.verify<JwtPayload>(token);
    } catch {
      throw new UnauthorizedException({
        success: false,
        error: 'INVALID_TOKEN',
        message: 'Invalid or expired token',
        meta: null,
      });
    }
  }
}
