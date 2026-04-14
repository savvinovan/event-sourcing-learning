package account

import (
	"github.com/savvinovan/wallet-service/internal/application/command"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

const (
	CmdOpenAccount     command.CommandType = "OpenAccount"
	CmdDepositMoney    command.CommandType = "DepositMoney"
	CmdWithdrawMoney   command.CommandType = "WithdrawMoney"
	CmdActivateAccount command.CommandType = "ActivateAccount"
	CmdFreezeAccount   command.CommandType = "FreezeAccount"
)

type OpenAccountCommand struct {
	AccountID  domain.AccountID
	CustomerID domain.CustomerID
	Currency   string
}

func (c OpenAccountCommand) CommandType() command.CommandType { return CmdOpenAccount }

type DepositMoneyCommand struct {
	AccountID domain.AccountID
	Amount    domain.Money
}

func (c DepositMoneyCommand) CommandType() command.CommandType { return CmdDepositMoney }

type WithdrawMoneyCommand struct {
	AccountID domain.AccountID
	Amount    domain.Money
}

func (c WithdrawMoneyCommand) CommandType() command.CommandType { return CmdWithdrawMoney }

type ActivateAccountCommand struct {
	AccountID domain.AccountID
}

func (c ActivateAccountCommand) CommandType() command.CommandType { return CmdActivateAccount }

type FreezeAccountCommand struct {
	AccountID domain.AccountID
	Reason    string
}

func (c FreezeAccountCommand) CommandType() command.CommandType { return CmdFreezeAccount }
