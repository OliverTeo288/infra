# IAM Permissions

The minimum permissions required for each `infra` command. 

Replace `<gitlab-host>` with your build's actual GitLab OIDC host (e.g.
the value baked into `config.GitlabHTTPSDomain`).

---

## `infra portforward`

Port-forwarding into a private RDS instance through ECS or EC2 over SSM.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds:DescribeDBInstances",
        "rds:DescribeDBProxies",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ecs:ListClusters",
        "ecs:ListServices",
        "ecs:ListTasks",
        "ecs:DescribeTasks"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ssm:StartSession",
      "Resource": [
        "arn:aws:ec2:*:*:instance/*",
        "arn:aws:ecs:*:*:task/*",
        "arn:aws:ssm:*:*:document/AWS-StartPortForwardingSessionToRemoteHost"
      ]
    }
  ]
}
```

## `infra ecs exec`

Executing shell commands interactively in ECS containers.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:ListClusters",
        "ecs:ListServices",
        "ecs:ListTasks",
        "ecs:DescribeTasks",
        "ecs:ExecuteCommand",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ssm:StartSession",
      "Resource": [
        "arn:aws:ecs:*:*:task/*",
        "arn:aws:ssm:*:*:document/AmazonECS-ExecuteInteractiveCommand"
      ]
    }
  ]
}
```

## `infra init`

Bootstrapping a project — template repo + S3 state bucket + GitOps IAM
role + OIDC provider.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:PutBucketVersioning",
        "s3:PutBucketPolicy",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::*-backend-tf-*",
        "arn:aws:s3:::*-backend-tf-*/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:PutRolePolicy",
        "iam:CreateOpenIDConnectProvider",
        "iam:GetOpenIDConnectProvider"
      ],
      "Resource": [
        "arn:aws:iam::*:role/TerraformGitopsRole",
        "arn:aws:iam::*:oidc-provider/<gitlab-host>"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:GetCallerIdentity",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    }
  ]
}
```

## `infra ecr read` / `infra ecr write`

Creating cross-account-assumable ECR roles.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:PutRolePolicy"
      ],
      "Resource": [
        "arn:aws:iam::*:role/ecrreader",
        "arn:aws:iam::*:role/ecrwriter"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:GetCallerIdentity",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    }
  ]
}
```
