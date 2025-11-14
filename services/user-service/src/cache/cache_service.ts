import { Injectable, Inject } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import type { Cache } from 'cache-manager';

/**
 * Cache Service for managing preference caching
 * Provides utilities for get, set, delete, and invalidate operations
 */
@Injectable()
export class CacheService {
  constructor(@Inject(CACHE_MANAGER) private cacheManager: Cache) {}

  /**
   * Get user preferences from cache
   * @param userId - The user ID
   * @returns Cached preferences or null if not found
   */
  async getUserPreferences(userId: string): Promise<any> {
    const cacheKey = this.getUserPreferencesCacheKey(userId);
    return await this.cacheManager.get(cacheKey);
  }

  /**
   * Set user preferences in cache
   * @param userId - The user ID
   * @param preferences - The preferences data to cache
   * @param ttl - Optional TTL in seconds (defaults to configured TTL)
   */
  async setUserPreferences(
    userId: string,
    preferences: any,
    ttl?: number,
  ): Promise<void> {
    const cacheKey = this.getUserPreferencesCacheKey(userId);
    await this.cacheManager.set(
      cacheKey,
      preferences,
      ttl ? ttl * 1000 : undefined,
    );
  }

  /**
   * Invalidate (delete) user preferences from cache
   * @param userId - The user ID
   */
  async invalidateUserPreferences(userId: string): Promise<void> {
    const cacheKey = this.getUserPreferencesCacheKey(userId);
    await this.cacheManager.del(cacheKey);
  }

  /**
   * Get multiple user preferences from cache
   * @param userIds - Array of user IDs
   * @returns Map of userId -> cached preferences (only for users found in cache)
   */
  async getBatchUserPreferences(
    userIds: string[],
  ): Promise<Map<string, unknown>> {
    const result = new Map<string, unknown>();

    await Promise.all(
      userIds.map(async (userId) => {
        const cached = (await this.getUserPreferences(userId)) as unknown;
        if (cached) {
          result.set(userId, cached);
        }
      }),
    );

    return result;
  }

  /**
   * Set multiple user preferences in cache
   * @param preferencesMap - Map of userId -> preferences data
   * @param ttl - Optional TTL in seconds
   */
  async setBatchUserPreferences(
    preferencesMap: Map<string, unknown>,
    ttl?: number,
  ): Promise<void> {
    await Promise.all(
      Array.from(preferencesMap.entries()).map(([userId, preferences]) =>
        this.setUserPreferences(userId, preferences, ttl),
      ),
    );
  }

  /**
   * Invalidate multiple user preferences from cache
   * @param userIds - Array of user IDs
   */
  async invalidateBatchUserPreferences(userIds: string[]): Promise<void> {
    await Promise.all(
      userIds.map((userId) => this.invalidateUserPreferences(userId)),
    );
  }

  /**
   * Clear all cache entries
   * Note: reset() is only available in memory store
   * For Redis, this method will be a no-op
   */
  async clearAll(): Promise<void> {
    // The store property might not have reset method in all implementations
    const manager = this.cacheManager as unknown as {
      reset?: () => Promise<void>;
    };
    if (typeof manager.reset === 'function') {
      await manager.reset();
    } else {
      console.warn('Cache store does not support reset operation');
    }
  }

  /**
   * Generate cache key for user preferences
   * @param userId - The user ID
   * @returns Cache key string
   */
  private getUserPreferencesCacheKey(userId: string): string {
    return `user:preferences:${userId}`;
  }

  /**
   * Get idempotent response from cache
   * @param requestId - The unique request ID
   * @returns Cached response or null if not found
   */
  async getIdempotentResponse(requestId: string): Promise<any> {
    const cacheKey = this.getIdempotencyKey(requestId);
    return await this.cacheManager.get(cacheKey);
  }

  /**
   * Set idempotent response in cache
   * @param requestId - The unique request ID
   * @param response - The response data to cache
   * @param ttl - TTL in seconds (default: 24 hours)
   */
  async setIdempotentResponse(
    requestId: string,
    response: any,
    ttl: number = 86400,
  ): Promise<void> {
    const cacheKey = this.getIdempotencyKey(requestId);
    await this.cacheManager.set(cacheKey, response, ttl * 1000);
  }

  /**
   * Generate cache key for idempotency
   * @param requestId - The unique request ID
   * @returns Cache key string
   */
  private getIdempotencyKey(requestId: string): string {
    return `idempotency:${requestId}`;
  }
}
