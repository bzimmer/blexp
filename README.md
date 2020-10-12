# Overview

Submit expenses from the command line

# Install
With an appropriately configured `GOPATH`:

$ go install

# Requirements

First, follow the [instructions](https://integrations.expensify.com/Integration-Server/doc/) for acquiring an API key. Once you have the key, create a file named `$HOME/.blexp.toml` with the following format:

```
NAME:
   blexp - submit expenses from the cli

USAGE:
   blexp [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE, -c FILE     Load configuration from FILE (default: $HOME/.blexp.toml)
   --force, -f                Force expense submission (default: false)
   --list, -l                 List expense templates (default: false)
   --expense value, -e value  Submit an expense for the named template
   --help, -h                 show help (default: false)

Acquire a UserID and UserSecret at https://integrations.expensify.com/Integration-Server/doc/

Possible expense template values can be found at https://integrations.expensify.com/Integration-Server/doc/#expense-creator
 - All names use SnakeCase

The configuration file format:

  UserID = "some_user_id"
  UserSecret = "lTLF3PjlcdhkcNgwUkPj2Q"
  UserEmail = "me@example"
  Default = "Broadband"

  [Templates]

  [Templates.Broadband]
  Merchant  = "Comcast"
  Amount    = 5000
  Currency  = "USD"
  Category  = "Employee Reimbursement"
  Tag       = "20 - My Department"

  [Templates.Lunch]
  Merchant  = "My Favorite Lunch Place"
  Amount    = 1500
  Currency  = "USD"
  Category  = "Entertainment"
  Tag       = "20 - My Department"
  ```

To submit expenses use the `-f` flag:

```sh
  ~/.../blexp (master) > dist/blexp -f
{"level":"info","op":"submit","exp":{"merchant":"Comcast","created":"2020-10-11","amount":5000,"currency":"USD","category":"Employee Reimbursement","tag":"20 - My Department","comment":"blexp: 3981e936"},"time":"2020-10-11T18:39:53-07:00","message":"submitting expense"}
{"level":"info","op":"submitted","exp":{"merchant":"Comcast","created":"2020-10-11","amount":5000,"currency":"USD","category":"Employee Reimbursement","tag":"20 - My Department","comment":"blexp: 3981e936","reportID":6509469,"transactionID":"4309220672839"},"time":"2020-10-11T18:39:54-07:00"}
```

Note a `reportID` & `transactionID` is reported if the successful submission of an expense.
