import {
  Entity,
  Column,
  PrimaryColumn,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

@Entity('simple_users')
export class SimpleUser {
  @PrimaryColumn({ type: 'varchar', length: 50 })
  user_id: string;

  @Column({ type: 'varchar', length: 255 })
  name: string;

  @Column({ type: 'varchar', length: 255, unique: true })
  email: string;

  @Column({ type: 'varchar', length: 255 })
  password: string;

  @Column({ type: 'text', nullable: true })
  push_token?: string;

  @Column({ type: 'boolean', default: true })
  email_preference: boolean;

  @Column({ type: 'boolean', default: true })
  push_preference: boolean;

  @Column({ type: 'timestamp', nullable: true })
  last_notification_email?: Date;

  @Column({ type: 'timestamp', nullable: true })
  last_notification_push?: Date;

  @Column({ type: 'varchar', length: 100, nullable: true })
  last_notification_id?: string;

  @CreateDateColumn()
  created_at: Date;

  @UpdateDateColumn()
  updated_at: Date;
}
