#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { ContentModerationInfraStack } from "../lib/infra-stack";

const app = new cdk.App();
new ContentModerationInfraStack(app, "InfraStack", {});
