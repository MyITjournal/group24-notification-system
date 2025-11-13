import { Module } from '@nestjs/common';
import { CacheModule as NestCacheModule } from '@nestjs/cache-manager';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { redisStore } from 'cache-manager-redis-yet';
import { CacheService } from './cache.service';

@Module({
  imports: [
    NestCacheModule.registerAsync({
      imports: [ConfigModule],
      inject: [ConfigService],
      useFactory: async (configService: ConfigService) => {
        const redisUrl = configService.get<string>('REDIS_URL');
        const ttl = configService.get<number>('CACHE_TTL', 3600); // Default 1 hour

        // If Redis URL is provided (production/Heroku), use Redis
        if (redisUrl) {
          return {
            store: await redisStore({
              url: redisUrl,
              ttl: ttl * 1000, // Convert to milliseconds
            }),
          };
        }

        // Otherwise, use in-memory cache (development)
        return {
          ttl: ttl * 1000, // Convert to milliseconds
          max: 100, // Maximum number of items in cache
        };
      },
    }),
  ],
  providers: [CacheService],
  exports: [CacheService, NestCacheModule],
})
export class CacheModule {}
