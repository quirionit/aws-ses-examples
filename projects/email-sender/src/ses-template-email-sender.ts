import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { EventBus, Rule } from 'aws-cdk-lib/aws-events';
import { SfnStateMachine } from 'aws-cdk-lib/aws-events-targets';
import { PolicyStatement, Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';
import { Architecture } from 'aws-cdk-lib/aws-lambda';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import { Choice, Condition, JsonPath, StateMachine, StateMachineType } from 'aws-cdk-lib/aws-stepfunctions';
import { CallAwsService, LambdaInvoke } from 'aws-cdk-lib/aws-stepfunctions-tasks';
import { Construct } from 'constructs';
import { EmailSenderConfiguration } from './email-sender-configuration';
import { LogGroup } from 'aws-cdk-lib/aws-logs';
import { RemovalPolicy } from 'aws-cdk-lib';
import { LogLevel } from 'aws-cdk-lib/aws-stepfunctions';

// eslint-disable-next-line @typescript-eslint/no-require-imports
const path = require('path');

export class SesTemplateEmailSender extends Construct {
  constructor(scope: Construct, id: string, props: EmailSenderConfiguration) {
    super(scope, id);

    /*--------Service definitions--------*/
    // Create a Lambda function that builds the raw MIME message
    const lambda = new GoFunction(this, `${id}-SesSendEmailLambda`, {
      functionName: `${id}-SesSendEmailLambda`,
      entry: path.resolve(__dirname, '../../go-src/sfn/send-email-with-attachment'),
      moduleDir: path.resolve(__dirname, '../../go-src/go.mod'),
      environment: {
        BUCKET: props.documentBucketName,
      },
      architecture: Architecture.ARM_64,
      bundling: {
        goBuildFlags: ['-ldflags "-s -w"'],
      },
    });

    // Add the required role policies to the Lambda function
    lambda.addToRolePolicy(new PolicyStatement({
      resources: [`arn:aws:s3:::${props.documentBucketName}/*`],
      actions: ['s3:GetObject'],
    }));
    lambda.addToRolePolicy(new PolicyStatement({
      resources: ['*'],
      actions: ['ses:SendRawEmail'],
    }));

    /*--------Step function tasks--------*/
    // Create a ses sendTemplatedEmail step function task
    const sendEmail = new CallAwsService(this, 'Send email', {
      service: 'sesv2',
      action: 'sendEmail',
      parameters: {
        Content: {
          Template: {
            TemplateData: JsonPath.objectAt('$.parameter'),
            TemplateName: JsonPath.objectAt('$.template')
          },
        },
        Destination: {
          ToAddresses: JsonPath.array(JsonPath.stringAt('$.recipient')),
        },
        FromEmailAddress: JsonPath.stringAt('$.sender'),
      },
      iamResources: ['*'],
      additionalIamStatements: [
        new PolicyStatement({
          resources: ['*'],
          actions: ['ses:SendTemplatedEmail'],
        }),
      ],
    });

    // Create a ses testRenderTemplate step function task
    const renderEmail = new CallAwsService(this, 'Render email', {
      service: 'sesv2',
      action: 'testRenderEmailTemplate',
      parameters: {
        TemplateName: JsonPath.objectAt('$.template'),
        TemplateData: JsonPath.objectAt('$.parameter'),
      },
      iamResources: ['*'],
      resultPath: '$.raw',
      additionalIamStatements: [
        new PolicyStatement({
          resources: ['*'],
          actions: ['ses:TestRenderEmailTemplate'],
        }),
      ],
    });

    // Create a step function task that invokes the lambda function
    const sendEmailWithAttachment = new LambdaInvoke(scope, 'Send email with attachment', {
      lambdaFunction: lambda,
    });

    /*--------Step function definition--------*/
    // Define the step function workflow
    const definition = new Choice(this, 'Has document', { inputPath: '$.detail' })
      .when(Condition.and(Condition.isPresent('$.documents'), Condition.isPresent('$.documents[0]')), renderEmail.next(sendEmailWithAttachment))
      .otherwise(sendEmail);

    // Create a step function state machine using the defined step function workflow
    const sfn = new StateMachine(scope, 'Send email state machine', {
      stateMachineName: `${id}-EmailSender`,
      stateMachineType: StateMachineType.EXPRESS,
      tracingEnabled: true,
      logs: {
        destination: new LogGroup(this, 'EmailWorkflowLogGroup', {
          logGroupName: 'EmailWorkflowLogGroup',
          removalPolicy: RemovalPolicy.DESTROY,
        }),
        level: LogLevel.ALL,
      },
      definition,
    });

    // grant the step function state machine the permissions to 


    /*--------Step function trigger--------*/
    // Create a new EventBridge rule for the provided EventBus
    const rule = new Rule(this, 'EmailSender_EventBusSubscriptionRule', {
      description: `Subscription to event bus for ${this.node.id}`,
      ruleName: 'EmailSender_Rule',
      eventPattern: props.eventPattern,
      eventBus: EventBus.fromEventBusName(this, 'EmailSender_EventBus', props.eventBusName),
    });

    // Add the created step function state machine as a target for the EventBridge rule
    rule.addTarget(new SfnStateMachine(sfn, {
      deadLetterQueue: new Queue(this, 'EmailSender_QueueDlq', {
        queueName: 'EmailSenderQueueDlq',
      }),
      role: new Role(this, `${this.node.id}EventRole`, {
        assumedBy: new ServicePrincipal('events.amazonaws.com'),
      }),
    }));
  }
}
