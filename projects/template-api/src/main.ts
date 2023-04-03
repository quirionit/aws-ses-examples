import { App, Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import {SesTemplateApi} from "./ses-template-api";

export class TemplateApiStack extends Stack {
  constructor(scope: Construct, id: string, props: StackProps = {}) {
    super(scope, id, props);
    new SesTemplateApi(this, 'TemplateApi')
  }
}

const devEnv = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION,
};

const app = new App();
new TemplateApiStack(app, 'ses-examples-dev', { env: devEnv });

app.synth();
