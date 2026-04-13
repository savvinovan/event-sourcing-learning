package account

import "github.com/savvinovan/wallet-service/internal/application/command"

const (
	CmdOpenAccount     command.CommandType = "OpenAccount"
	CmdDepositMoney    command.CommandType = "DepositMoney"
	CmdWithdrawMoney   command.CommandType = "WithdrawMoney"
	CmdActivateAccount command.CommandType = "ActivateAccount"
	CmdFreezeAccount   command.CommandType = "FreezeAccount"
)

type OpenAccountCommand struct {
	AccountID  string
	CustomerID string
	Currency   string
}

func (c OpenAccountCommand) CommandType() command.CommandType { return CmdOpenAccount }

type DepositMoneyCommand struct {
	AccountID string
	Amount    int64
	Currency  string
}

func (c DepositMoneyCommand) CommandType() command.CommandType { return CmdDepositMoney }

type WithdrawMoneyCommand struct {
	AccountID string
	Amount    int64
	Currency  string
}

func (c WithdrawMoneyCommand) CommandType() command.CommandType { return CmdWithdrawMoney }

type ActivateAccountCommand struct {
	AccountID string
}

func (c ActivateAccountCommand) CommandType() command.CommandType { return CmdActivateAccount }

type FreezeAccountCommand struct {
	AccountID string
	Reason    string
}

func (c FreezeAccountCommand) CommandType() command.CommandType { return CmdFreezeAccount }
