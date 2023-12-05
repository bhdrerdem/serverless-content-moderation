import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as sns from "aws-cdk-lib/aws-sns";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as iam from "aws-cdk-lib/aws-iam";
import * as lambdaEventSources from "aws-cdk-lib/aws-lambda-event-sources";
import { SqsSubscription } from "aws-cdk-lib/aws-sns-subscriptions";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";

export class ContentModerationInfraStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const queue = new sqs.Queue(this, "ContentModerationQueue", {
            visibilityTimeout: cdk.Duration.seconds(300),
        });

        const topic = new sns.Topic(this, "ContentModerationTopic");
        topic.addSubscription(
            new SqsSubscription(queue, {
                rawMessageDelivery: true,
            })
        );

        const table = new dynamodb.Table(this, "ContentModerationTable", {
            partitionKey: { name: "id", type: dynamodb.AttributeType.STRING },
        });

        const serverLambda = new lambda.Function(this, "ServerLambda", {
            runtime: lambda.Runtime.GO_1_X,
            handler: "main",
            code: lambda.Code.fromAsset("../app/bin/api"),
            environment: {
                CONTENT_MODERATION_TOPIC_ARN: topic.topicArn,
                GIN_MODE: "release",
            },
        });

        const api = new RestApi(this, "ContentModerationApi", {
            restApiName: "ContentModerationRest",
        });

        const lambdaIntegration = new LambdaIntegration(serverLambda);

        const postResource = api.root;
        postResource.addMethod("POST", lambdaIntegration);

        const sqsConsumerLambda = new lambda.Function(
            this,
            "SqsConsumerLambda",
            {
                runtime: lambda.Runtime.GO_1_X,
                handler: "main",
                code: lambda.Code.fromAsset("../app/bin/consume"),
                environment: {
                    CONTENT_MODERATION_TABLE_NAME: table.tableName,
                },
            }
        );

        topic.grantPublish(serverLambda);

        table.grantReadWriteData(sqsConsumerLambda);
        queue.grantConsumeMessages(sqsConsumerLambda);

        sqsConsumerLambda.addEventSource(
            new lambdaEventSources.SqsEventSource(queue)
        );

        const comprehendPolicy = new iam.PolicyStatement({
            actions: ["comprehend:Detect*"],
            resources: ["*"],
        });

        sqsConsumerLambda.role?.attachInlinePolicy(
            new iam.Policy(this, "ComprehendPolicy", {
                statements: [comprehendPolicy],
            })
        );
    }
}
