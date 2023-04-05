package view

import "github.com/AlecAivazis/survey/v2"

var (
	unauthorizedMenuItems = []string{login, register, quite}
	authorizedMenuItems   = []string{logout, addSecret, getSecret, deleteSecret, listSecrets, synchronize, quite}
)

func getMainMenu(isAuthorized bool) *survey.Select {
	var menuOptions []string
	if isAuthorized {
		menuOptions = authorizedMenuItems
	} else {
		menuOptions = unauthorizedMenuItems
	}

	return &survey.Select{
		Message: "What do you want ?:",
		Options: menuOptions,
	}
}

var loginQuestions = []*survey.Question{
	{
		Name:     "Login",
		Prompt:   &survey.Input{Message: "Enter your Login"},
		Validate: survey.Required,
	},
	{
		Name:     "Password",
		Prompt:   &survey.Password{Message: "Enter your Password"},
		Validate: survey.Required,
	},
}

type loginAnswer struct {
	Login    string
	Password string
}

var registerQuestions = []*survey.Question{
	{
		Name:     "Login",
		Prompt:   &survey.Input{Message: "Enter your Login"},
		Validate: survey.MinLength(3),
	},
	{
		Name:     "Password",
		Prompt:   &survey.Password{Message: "Enter your Password"},
		Validate: survey.Required,
	},
	{
		Name:     "RepeatedPassword",
		Prompt:   &survey.Password{Message: "Repeat Password"},
		Validate: survey.Required,
	},
}

type resisterAnswer struct {
	Login            string
	Password         string
	RepeatedPassword string
}

var addCredentialsQuestions = []*survey.Question{
	{
		Name:     "Name",
		Prompt:   &survey.Input{Message: "Enter secret name to store"},
		Validate: survey.MinLength(3),
	},
	{
		Name:     "Login",
		Prompt:   &survey.Input{Message: "Enter your Login to store"},
		Validate: survey.Required,
	},
	{
		Name:     "Password",
		Prompt:   &survey.Password{Message: "Enter your Password to store"},
		Validate: survey.Required,
	},
	{
		Name:     "Description",
		Prompt:   &survey.Input{Message: "Enter description to store"},
		Validate: survey.Required,
	},
}

type addCredentialsAnswer struct {
	Name        string
	Login       string
	Password    string
	Description string
}

var addTextQuestions = []*survey.Question{
	{
		Name:     "Name",
		Prompt:   &survey.Input{Message: "Enter secret name"},
		Validate: survey.MinLength(3),
	},
	{
		Name:     "Text",
		Prompt:   &survey.Input{Message: "Enter your textual data"},
		Validate: survey.Required,
	},
	{
		Name:     "Description",
		Prompt:   &survey.Input{Message: "Enter description"},
		Validate: survey.Required,
	},
}

type addTextAnswer struct {
	Name        string
	Text        string
	Description string
}

var addBinaryQuestions = []*survey.Question{
	{
		Name:     "Name",
		Prompt:   &survey.Input{Message: "Enter secret name"},
		Validate: survey.MinLength(3),
	},
	{
		Name:     "Filepath",
		Prompt:   &survey.Input{Message: "Enter filepath"},
		Validate: survey.Required,
	},
	{
		Name:     "Description",
		Prompt:   &survey.Input{Message: "Enter description"},
		Validate: survey.Required,
	},
}

type addBinaryAnswer struct {
	Name        string
	FilePath    string
	Description string
}

var addCardQuestions = []*survey.Question{
	{
		Name:     "Name",
		Prompt:   &survey.Input{Message: "Enter secret name"},
		Validate: survey.MinLength(3),
	},
	{
		Name:     "CardNumber",
		Prompt:   &survey.Input{Message: "Enter card number"},
		Validate: survey.Required,
	},
	{
		Name:     "CardName",
		Prompt:   &survey.Input{Message: "Enter card owner name"},
		Validate: survey.Required,
	},
	{
		Name:     "CardCVV",
		Prompt:   &survey.Input{Message: "Enter cvv"},
		Validate: survey.Required,
	},
	{
		Name:     "Description",
		Prompt:   &survey.Input{Message: "Enter description"},
		Validate: survey.Required,
	},
}

type addCardAnswer struct {
	Name        string
	CardNumber  string
	CardName    string
	CardCVV     string
	Description string
}

var addSelectOptions = &survey.Select{
	Message: "What type of credentials do you want to save ?:",
	Options: []string{credentials, text, binary, card},
}

var getSecretNameQuestion = &survey.Input{
	Message: "Enter stored secret name:",
}

var getStoreFilepathQuestion = &survey.Input{
	Message: "Enter the path to the folder where you want to save the file: ",
}
