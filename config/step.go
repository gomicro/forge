package config

// Step represents details of single step to be executed by the cli.
type Step struct {
	Cmd Cmd `yaml:"cmd"`
	//Cmds []string `yaml:"cmds"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute() (string, error) {
	return s.Cmd.Execute()
}
