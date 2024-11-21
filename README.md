# infra

Easing your AWS painpoints by:

1) Portforwarding into private RDS through ECS Fargate through SSM.
2) Initialisation of repository by templated SHIPHATS GitLab Terraform templates


## Install via Homebrew

```bash
$ brew tap oliverteo288/infra

$ brew install oliverteo288/infra/infra
```

## Usage

````bash
## By utilising your AWS profile, it will discovers available ECS to 
infra portforward

## Ensure have SHIPHATS permissions, it will clone templated repository onto local folder
infra init
````  