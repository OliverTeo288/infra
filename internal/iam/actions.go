package iam

import "slices"

// ECRReadActions is the least-privilege set of ECR API actions an image puller
// needs (pull, describe, list).
var ECRReadActions = []string{
	"ecr:BatchCheckLayerAvailability",
	"ecr:GetDownloadUrlForLayer",
	"ecr:GetRepositoryPolicy",
	"ecr:DescribeRepositories",
	"ecr:ListImages",
	"ecr:DescribeImages",
	"ecr:BatchGetImage",
	"ecr:GetLifecyclePolicy",
	"ecr:GetLifecyclePolicyPreview",
	"ecr:ListTagsForResource",
	"ecr:DescribeImageScanFindings",
}

// ECRWriteActions adds image-push actions on top of the read set.
var ECRWriteActions = slices.Concat(ECRReadActions, []string{
	"ecr:PutImage",
	"ecr:InitiateLayerUpload",
	"ecr:UploadLayerPart",
	"ecr:CompleteLayerUpload",
})
