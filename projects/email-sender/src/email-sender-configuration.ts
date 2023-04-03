import { EventPattern } from 'aws-cdk-lib/aws-events/lib/event-pattern';

export interface EmailSenderConfiguration {
  /**
   * Eventbus name to trigger ses
   */
  readonly eventBusName: string;

  /**
   * Event pattern to trigger ses
   */
  readonly eventPattern: EventPattern;

  /**
   * S3 bucket name where the documents are stored
   */
  readonly documentBucketName: string;
}
