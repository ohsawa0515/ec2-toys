ec2-toys
===

![Test Status](https://github.com/ohsawa0515/ec2-toys/actions/workflows/test.yml/badge.svg)


Useful cli tools for Amazon EC2.

* [How to install and settings](#how-to-install-and-settings)
  * [1\. Installation](#1-installation)
  * [2\. Set AWS credentials](#2-set-aws-credentials)
  * [3\. Set AWS region](#3-set-aws-region)
* [Usage](#usage)
  * [Run](#run)
* [Options](#options)
  * [Global](#global)
    * [\-\-region (\-r)](#--region--r)
    * [\-\-profile (\-p)](#--profile--p)
  * [List](#list)
    * [\-\-filters (\-f)](#--filters--f)
* [Combination example of the peco command](#combination-example-of-the-peco-command)
  * [SSH public EC2 instance\.](#ssh-public-ec2-instance)
  * [Via bastion server\.](#via-bastion-server)
* [Contribution](#contribution)
* [License](#license)

## How to install and settings

### 1. Installation

```
$ go get -u github.com/ohsawa0515/ec2-toys
```

### 2. Set AWS credentials

Please ignore if you use [IAM Roles](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html).

* Credential file (`$HOME/.aws/credentials`) 

```
[default]
aws_access_key_id = <YOUR_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
```

If you want to use [profile](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-multiple-profiles), and set as follows.

```
[default]
aws_access_key_id = <YOUR_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>

[dev]
aws_access_key_id = <YOUR_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
```

* Environment variable

```
$ export AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY_ID
$ export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY
```

### 3. Set AWS region

* shared config (`$HOME/.aws/config`) 

```
[default]
region = us-east-1
```

If you want to use [profile](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-multiple-profiles), and set as follows.

```
[default]
region = us-east-1

[dev]
region = us-east-1
```

* Environment variable

```
$ export AWS_REGION=us-east-1
```

## Usage

### Run

After run `ec2-toys list`, listing ec2 instances.

```bash
$ ec2-toys list
admin-server    192.0.2.11  203.0.113.11  i-xxxxxxxx    c3.large    us-east-1c  running linux
batch-server001 192.0.2.20  203.0.113.20  i-xxxxxxxx    t2.medium   us-east-1c  running windows
web-server001   192.0.2.12  203.0.113.12  i-xxxxxxxx    t2.micro    us-east-1a  running linux
web-server002   192.0.2.13  203.0.113.13  i-xxxxxxxx    t2.medium   us-east-1c  stopped linux
```

Support `--filter` option like [describe-instances command](http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html).

```bash
# Running instances.
$ ec2-toys list --filters "Name=instance-state-name,Values=running"
admin-server    192.0.2.11  203.0.113.11  i-xxxxxxxx    c3.large    us-east-1c  running linux
batch-server001 192.0.2.20  203.0.113.20  i-xxxxxxxx    t2.medium   us-east-1c  running windows
web-server001   192.0.2.12  203.0.113.12  i-xxxxxxxx    t2.micro    us-east-1a  running linux

# Instance type is t2.medium
$ ec2-toys list --filters "Name=instance-type,Values=t2.medium"
batch-server001 192.0.2.20  203.0.113.20  i-xxxxxxxx    t2.medium   us-east-1c  running windows
web-server002   192.0.2.13  203.0.113.13  i-xxxxxxxx    t2.medium   us-east-1c  stopped linux

# Running and t2.medium instances(Separate by space).
$ ec2-toys list --filters "Name=instance-state-name,Values=running Name=instance-type,Values=t2.medium"
batch-server001 192.0.2.20  203.0.113.20  i-xxxxxxxx    t2.medium   us-east-1c  running windows
```

## Options

### Global

#### --region (-r)

The region to use. Overrides config/env settings.

e.g.

```
$ ec2-toys list --region ap-northeast-1  # Tokyo region
```

#### --profile (-p)

Use a specific profile from your credential file.
See [cli-multiple-profiles](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-multiple-profiles).

e.g.

```
# If you have the following settings,
[dev]
aws_access_key_id = <YOUR_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>

# You use profile dev.
$ ec2-toys list --profile dev
```

### List

#### --filters (-f)

Filtering EC2 instances.

e.g.

```
# Running instances.
$ ec2-toys list --filters "Name=instance-state-name,Values=running"

# Instance type is t2.medium
$ ec2-toys list --filters "Name=instance-type,Values=t2.medium"

# Running and t2.medium instances(Separate by space).
$ ec2-toys list --filters "Name=instance-state-name,Values=running Name=instance-type,Values=t2.medium"

# Specify VPC ID
$ ec2-toys list --filters "Name=vpc-id,Values=vpc-xxxxxxxx"

# Filter instances with a Env=prod tag
$ ec2-toys list --filters "Name=tag:Env,Values=prod"
```

## Combination example of the peco command

[peco](https://github.com/peco/peco) is Simplistic interactive filtering tool. 
Combined with peco command, you cloud be conveniently SSH.

### SSH public EC2 instance.

```
$ alias ssh-ec2="ec2-toys list --filters \"Name=instance-state-name,Values=running\" | grep linux | peco | awk {'print \$3'} | xargs -I{} sh -c 'ssh ec2-user@{} < /dev/tty' ssh"
```

### Via bastion server.

Refer to http://qiita.com/kawaz/items/a0151d3aa2b6f9c4b3b8 .

```
# ~/.ssh/config

Host bastion-server/*
  ProxyCommand ssh -W "$(basename "%h")":%p "$(dirname "%h")"

Host bastion-sever
  Hostname xxx.xxx.xxx.xxx
  User foo
  IdentityFile ~/.ssh/id_rsa
```

```
$ alias ssh-bastion="ec2-toys list --filters \"Name=instance-state-name,Values=running\" | grep linux | peco | awk {'print \$2'} | xargs -I{} sh -c 'ssh bastion-server/{} < /dev/tty' ssh"
```

## Contribution

1. Fork ([https://github.com/ohsawa0515/ec2-toys/fork](https://github.com/ohsawa0515/ec2-toys/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `go test ./...` command and confirm that it passes
6. Run `gofmt -s`
7. Create new Pull Request

## License

See [LICENSE](https://github.com/ohsawa0515/ec2-toys/blob/master/LICENSE).
