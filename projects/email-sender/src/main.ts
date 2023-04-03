import {App, RemovalPolicy, Stack, StackProps} from 'aws-cdk-lib';
import { Construct } from 'constructs';
import {EventBus} from "aws-cdk-lib/aws-events";
import {Bucket} from "aws-cdk-lib/aws-s3";
import {SesTemplateEmailSender} from "./ses-template-email-sender";
const { createHash } = require('crypto');

export class EmailSenderStack extends Stack {
    constructor(scope: Construct, id: string, props: StackProps = {}) {
        super(scope, id, props);

        // Create an EventBridge EventBus
        const eventBus = new EventBus(this, 'EmailSenderEventBus', {
            eventBusName: 'emailSenderEventBus',
        });

        const bucketId = createHash('md5').update(this.account).digest('hex');

        // Create a S3 bucket to store the documents to be attached to the emails
        const bucket = new Bucket(this, `${id}-SesDocumentBucket`, {
            versioned: false,
            bucketName: `ses-documents-${bucketId}`,
            removalPolicy: RemovalPolicy.DESTROY,
            autoDeleteObjects: true,
        });

        // Create a bundle of constructs to simplify sending emails with SES
        new SesTemplateEmailSender(this, 'EmailSender', {
            eventBusName: eventBus.eventBusName,
            eventPattern: {
                source: ['example.email-sender'],
                detailType: ['template.send-email'],
            },
            documentBucketName: bucket.bucketName,
        });
    }
}

const devEnv = {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
};

const app = new App();
new EmailSenderStack(app, 'email-sender-dev', { env: devEnv });

app.synth();
