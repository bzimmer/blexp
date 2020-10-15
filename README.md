# Overview

Submit expenses from the command line

# Install
With an appropriately configured `GOPATH`:

```sh
$ task install
```

# Requirements

First, follow the [instructions](https://integrations.expensify.com/Integration-Server/doc/) for acquiring an API key. Once you have the key, create a file named `$HOME/.blexp.json` with the following format:

```
NAME:
   blexp - submit expenses from the cli

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   list, l    List expense templates
   submit, s  Submit expenses
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
  ```

The configuration file format:

```json
{
    "user_id": "** user_id **",
    "user_secret": "** user_secret_key **",
    "user_email": "me@example.com",
    "default": "Broadband",
    "templates": {
        "Broadband": {
            "merchant": "Comcast",
            "amount": 500,
            "currency": "USD",
            "category": "Employee Reimbursement",
            "tag": "30 - People and Me"
        },
        "Lunch": {
            "merchant": "My Favorite Lunch Place",
            "amount": 1500,
            "currency": "USD",
            "category": "Entertainment",
            "tag": "20 - My Department"
        }
    }
}
```

To list expenses use the `list` command:

```sh
~ > blexp list
8:41PM INF reading configuration path=/Users/<someone>/.blexp.json
8:41PM INF list default=true name=Broadband template={"amount":500,"category":"Employee Reimbursement","created":null,"currency":"USD","merchant":"Comcast","tag":"30 - People and Me"}
8:41PM INF list default=false name=Lunch template={"amount":1500,"category":"Entertainment","created":null,"currency":"USD","merchant":"My Favorite Lunch Place","tag":"20 - My Department"}
```

To test submit expenses use the `submit` command:

```sh
~ > blexp submit Lunch
8:43PM INF reading configuration path=/Users/<someone>/.blexp.json
8:43PM WRN submit exp={"amount":1500,"category":"Entertainment","created":null,"currency":"USD","merchant":"My Favorite Lunch Place","tag":"20 - My Department"} submitted=false template=Lunch
```

To submit expenses use the `submit` command with the `-f` flag:

```sh
~ > blexp submit -f Lunch
8:45PM INF reading configuration path=/Users/<someone>/.blexp.json
8:45PM INF submitting expense exp={"amount":1500,"category":"Entertainment","comment":"blexp: 55708577","created":"2020-10-14","currency":"USD","merchant":"My Favorite Lunch Place","tag":"20 - My Department"} op=submit
8:45PM INF submit exp={"amount":1500,"category":"Entertainment","comment":"blexp: 55708577","created":"2020-10-14","currency":"USD","merchant":"My Favorite Lunch Place","reportID":65918812,"tag":"20 - My Department","transactionID":"43095306711072841"} submitted=true template=Lunch
```

Note a `reportID` & `transactionID` are reported if the expense submission was successful.
