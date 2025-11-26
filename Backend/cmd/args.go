package cmd

import "flag"

type ProgramArgs struct {
	Cleanup bool
}

func GetProgramArgs() (*ProgramArgs, error) {
	args := &ProgramArgs{}
	flag.BoolVar(&args.Cleanup, "cleanup", false, "activate cleanup feature")
	flag.Parse()

	return args, nil
}
