# @quirionit/ses-examples

*@quirionit/aws-ses-examples* provides examples for an extensions to AWS SES such as an API to create nested templates and an extension to send attachments.

## Getting Started

For using the examples clone the repository
```bash
$ git clone git@github.com:quirionit/aws-ses-examples.git
$ cd aws-ses-examples/projects/go-src
$ go mod tidy
```

for *SesTemplateApi*
```bash
$ cd aws-ses-examples/projects/template-api
$ npm install
$ cdk deploy
```

for *SesTemplateEmailSender*
```bash
$ cd aws-ses-examples/projects/email-sender
$ npm install
$ cdk deploy
```

to delete the created services
```bash
$ cdk destroy
```
