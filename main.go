package main

import (
	"github.com/ohsawa0515/ec2-toys/aws-ec2"
	"gopkg.in/urfave/cli.v1"
	"os"
)

var BaseFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "region, r",
		Usage: "The region to use. Overrides config/env settings.",
	},
	cli.StringFlag{
		Name:  "profile, p",
		Usage: "Use a specific profile from your credential file.",
	},
}

var Commands = []cli.Command{
	commandInit,
}

var flagsInit = append(BaseFlags, cli.StringFlag{
	Name:  "filters, f",
	Usage: "Filtering ec2 tag",
})

var commandInit = cli.Command{
	Name:  "list",
	Usage: "List EC2 instances.",
	Action: func(c *cli.Context) error {
		instances, err := aws_ec2.DescribeInstances(c.String("region"), c.String("profile"), c.String("filters"))
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		aws_ec2.PrintInstances(instances)
		return nil
	},
	Flags: flagsInit,
}

func main() {
	app := cli.NewApp()
	app.Name = "ec2-toys"
	app.Usage = "Useful cli to operation Amazon EC2."
	app.Author = "Shuichi Ohsawa"
	app.Email = "ohsawa0515@gmail.com"
	app.Version = "0.0.1"
	app.Flags = BaseFlags
	app.Commands = Commands
	app.Run(os.Args)
}
