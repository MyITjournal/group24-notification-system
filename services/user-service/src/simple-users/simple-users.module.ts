import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { SimpleUsersController } from './simple-users.controller';
import { SimpleUsersService } from './simple-users.service';
import { SimpleUser } from './entity/simple-user.entity';
import { CacheModule } from '../cache/cache.module';

@Module({
  imports: [TypeOrmModule.forFeature([SimpleUser]), CacheModule],
  controllers: [SimpleUsersController],
  providers: [SimpleUsersService],
  exports: [SimpleUsersService],
})
export class SimpleUsersModule {}
