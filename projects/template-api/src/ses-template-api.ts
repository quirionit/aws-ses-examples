import { createHash } from 'crypto';
import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { RemovalPolicy, Stack } from 'aws-cdk-lib';
import { AttributeType, Table } from 'aws-cdk-lib/aws-dynamodb';
import { PolicyStatement } from 'aws-cdk-lib/aws-iam';
import { Architecture } from 'aws-cdk-lib/aws-lambda';
import { Bucket } from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';

// eslint-disable-next-line @typescript-eslint/no-require-imports
const path = require('path');

export class SesTemplateApi extends Construct {
  public readonly bucket: Bucket;

  constructor(scope: Construct, id: string) {
    super(scope, id);

    const bucketId = createHash('md5').update(Stack.of(this).account).update(id).digest('hex');

    // Create a S3 bucket to store the html template files
    this.bucket = new Bucket(this, `${id}-SesTemplateBucket`, {
      versioned: false,
      bucketName: `ses-email-templates-${bucketId}`,
      removalPolicy: RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
    });

    // Create a DynamoDB table to store a template's metadata
    const table = new Table(this, `${id}-SesTemplateTable`, {
      partitionKey: { name: 'TemplateName', type: AttributeType.STRING },
      removalPolicy: RemovalPolicy.DESTROY,
    });

    // Create a GO Lambda function
    const lambda = new GoFunction(this, `${id}-SesTemplateLambda`, {
      functionName: `${id}-SesTemplateLambda`,
      entry: path.resolve(__dirname, '../../go-src/api/'),
      moduleDir: path.resolve(__dirname, '../../go-src/go.mod'),
      environment: {
        DATA_STORE_TABLE: table.tableName,
        BUCKET: this.bucket.bucketName,
      },
      architecture: Architecture.ARM_64,
      bundling: {
        goBuildFlags: ['-ldflags "-s -w"'],
      },
    });

    // Add the required role policies to the Lambda function
    lambda.addToRolePolicy(new PolicyStatement({
      resources: [table.tableArn],
      actions: [
        'dynamodb:GetItem',
        'dynamodb:PutItem',
        'dynamodb:Scan',
      ],
    }));
    lambda.addToRolePolicy(new PolicyStatement({
      resources: [`${this.bucket.bucketArn}/*`],
      actions: [
        's3:GetObject',
        's3:PutObject',
      ],
    }));
    lambda.addToRolePolicy(new PolicyStatement({
      resources: ['*'],
      actions: ['ses:CreateEmailTemplate'],
    }));

    // Create an API Gateway and connect it to the Lambda function
    const integration = new HttpLambdaIntegration(`${id}-SesTemplateApiIntegration`, lambda);
    const api = new HttpApi(this, `${id}-SesTemplateApi`);
    api.addRoutes({
      path: '/emails',
      methods: [HttpMethod.POST],
      integration: integration,
    });
  }
}
