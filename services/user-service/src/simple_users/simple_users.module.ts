import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { SimpleUsersController } from './simple_users.controller';
import { SimpleUsersService } from './simple_users.service';
import { SimpleUser } from './entity/simple_user.entity';
import { CacheModule } from '../cache/cache_module';

@Module({
  imports: [TypeOrmModule.forFeature([SimpleUser]), CacheModule],
  controllers: [SimpleUsersController],
  providers: [SimpleUsersService],
  exports: [SimpleUsersService],
})
export class SimpleUsersModule {}
