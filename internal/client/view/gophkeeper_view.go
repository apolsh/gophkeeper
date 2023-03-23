package view

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"github.com/pterm/pterm"
)

const (
	login        string = "login"
	register     string = "register"
	logout       string = "logout"
	addSecret    string = "add secret"
	getSecret    string = "get secret"
	deleteSecret string = "delete secret"
	listSecrets  string = "list secrets"
	synchronize  string = "synchronize with remote"
	quite        string = "quite"
)

const (
	credentials string = "credentials"
	text        string = "text"
	binary      string = "binary"
	card        string = "card"
)

// GophkeeperViewInteractiveCLI cli implementation of IGophkeeperView
type GophkeeperViewInteractiveCLI struct {
	c            *controller.GophkeeperController
	isAuthorized bool
}

var _ controller.IGophkeeperView = (*GophkeeperViewInteractiveCLI)(nil)

// SetController sets controller for UI implementation
func (v *GophkeeperViewInteractiveCLI) SetController(controller *controller.GophkeeperController) {
	v.c = controller
}

// Show shows UI implementation
func (v *GophkeeperViewInteractiveCLI) Show(ctx context.Context) error {
MENU:
	for {
		select {
		case <-ctx.Done():
			break MENU
		default:
		}

		var variant string
		err := survey.AskOne(getMainMenu(v.isAuthorized), &variant, survey.WithValidator(survey.Required))
		if err != nil {
			fmt.Println(err)
			if err == terminal.InterruptErr {
				break MENU
			}
		}
		switch variant {
		case login:
			answers := loginAnswer{}
			err := survey.Ask(loginQuestions, &answers)
			if err != nil {
				fmt.Println(err)
				if err == terminal.InterruptErr {
					break MENU
				}
			}
			v.c.Login(ctx, answers.Login, answers.Password)
		case register:
			ans := resisterAnswer{}
			err := survey.Ask(registerQuestions, &ans)
			if err != nil {
				fmt.Println(err)
				if err == terminal.InterruptErr {
					break MENU
				}
			}
			v.c.Register(ctx, ans.Login, ans.Password, ans.RepeatedPassword)
		case logout:
			v.c.UnAuthorize()
		case addSecret:
			err := survey.AskOne(addSelectOptions, &variant, survey.WithValidator(survey.Required))
			if err != nil {
				fmt.Println(err)
				if err == terminal.InterruptErr {
					break MENU
				}
			}
			var secret model.SecretItem
			switch variant {
			case credentials:
				ans := addCredentialsAnswer{}
				err := survey.Ask(addCredentialsQuestions, &ans)
				if err != nil {
					fmt.Println(err)
					if err == terminal.InterruptErr {
						break MENU
					}
				}
				secret = model.NewCredentialsSecretItem(ans.Name, ans.Description, ans.Login, ans.Password)
			case text:
				ans := addTextAnswer{}
				err := survey.Ask(addTextQuestions, &ans)
				if err != nil {
					fmt.Println(err)
					if err == terminal.InterruptErr {
						break MENU
					}
				}
				secret = model.NewTextSecretItem(ans.Name, ans.Description, ans.Text)
			case binary:
				ans := addBinaryAnswer{}
				err := survey.Ask(addBinaryQuestions, &ans)
				if err != nil {
					fmt.Println(err)
					if err == terminal.InterruptErr {
						break MENU
					}
				}
				secret, err = model.NewBinarySecretItem(ans.Name, ans.Description, ans.FilePath)
				if err != nil {
					v.ShowError(err)
					continue
				}
			case card:
				ans := addCardAnswer{}
				err := survey.Ask(addCardQuestions, &ans)
				if err != nil {
					fmt.Println(err)
					if err == terminal.InterruptErr {
						break MENU
					}
				}
				secret = model.NewCardSecretItem(ans.Name, ans.Description, ans.CardName, ans.CardNumber, ans.CardCVV)
			}
			v.c.SaveSecret(ctx, secret)
		case getSecret:
			var name string
			err := survey.AskOne(getSecretNameQuestion, &name, survey.WithValidator(survey.Required))
			if err != nil {
				fmt.Println(err)
				if err == terminal.InterruptErr {
					break MENU
				}
			}
			v.c.GetSecret(ctx, name)
		case deleteSecret:
			var name string
			err := survey.AskOne(getSecretNameQuestion, &name, survey.WithValidator(survey.Required))
			if err != nil {
				fmt.Println(err)
				if err == terminal.InterruptErr {
					break MENU
				}
			}
			v.c.DeleteSecret(ctx, name)
		case listSecrets:
			v.c.ListSecret(ctx)
		case synchronize:
			v.c.Synchronize(ctx)
		case quite:
			fmt.Println("shutting down...")
			break MENU
		default:
			fmt.Println("Unknown command, try again please.")
		}
	}
	return nil
}

// SetAuthorized sets that user is already authorized for UI implementation
func (v *GophkeeperViewInteractiveCLI) SetAuthorized(isAuthorized bool) {
	v.isAuthorized = isAuthorized
}

// ViewSecretsInfoList shows secret info list
func (v *GophkeeperViewInteractiveCLI) ViewSecretsInfoList(secretInfos []dto.SecretItemInfo) {
	headers := []string{"NAME", "TYPE", "DESCRIPTION"}
	tableData := make(pterm.TableData, len(secretInfos)+1, len(secretInfos)+1)
	tableData = append(tableData, headers)
	for _, secretInfo := range secretInfos {
		row := []string{secretInfo.Name, secretInfo.SecretType, secretInfo.Description}
		tableData = append(tableData, row)
	}
	err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	if err != nil {
		v.ShowError(fmt.Errorf("failed to show secrets info: %w", err))
	}
}

// ShowSecretItem shows secret item
func (v *GophkeeperViewInteractiveCLI) ShowSecretItem(item model.SecretItem) {
	pterm.Info.Println(item.GetSecretPayload())
}

// ShowError shows error
func (v *GophkeeperViewInteractiveCLI) ShowError(err error) {
	pterm.Error.Println(err)
}

// GetStringInput gets input
func (v *GophkeeperViewInteractiveCLI) GetStringInput(ctx context.Context, inputText string) string {
	var path string
	err := survey.AskOne(getStoreFilepathQuestion, &path, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println(err)
		if err == terminal.InterruptErr {
			return ""
		}
		v.ShowError(err)
	}
	return path
}
