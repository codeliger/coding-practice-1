package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Account reprents a bank account
type Account struct {
	CustomerID             string
	Balance                float64
	DailyDepositLimit      float64
	WeeklyDepositLimit     float64
	DailyDepositVelocity   float64
	WeeklyDepositVelocity  float64
	DailyDepositCountLimit int
	DailyDepositCount      int
	LastDeposit            Transaction
}

// Transaction is a single transaction can potentially be a deposit or withdrawl its not specified
type Transaction struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	RawAmount  string `json:"load_amount"`
	Amount     float64
	Time       time.Time `json:"time"`
}

type Test struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Accepted   bool   `json:"accepted"`
}

// Deposit attempts to deposit an amount based on that users daily and weekly limits
// returns false on failure
func (account *Account) Deposit(t Transaction) bool {

	if (account.LastDeposit != Transaction{}) {
		// we are in the same year and month which means we can be the same day or same week
		_, lastDepositWeek := account.LastDeposit.Time.ISOWeek()
		_, transactionWeek := t.Time.ISOWeek()

		if account.LastDeposit.Time.Year() != t.Time.Year() || account.LastDeposit.Time.Month() != t.Time.Month() || lastDepositWeek != transactionWeek {
			// the weeks have changed so reset weekly and daily limits
			log.Println(account.CustomerID, "Year, month, or week changed resetting weekly and daily velocity")
			account.WeeklyDepositVelocity = 0
			account.DailyDepositVelocity = 0
			account.DailyDepositCount = 0
		} else if account.LastDeposit.Time.Day() != t.Time.Day() {
			log.Println(account.CustomerID, "Day changed resetting daily velocity")
			// the day of the week has changed so reset daily limits
			account.DailyDepositVelocity = 0
			account.DailyDepositCount = 0
		}
	}

	if account.WeeklyDepositVelocity+t.Amount > account.WeeklyDepositLimit {
		log.Println("Weekly Deposit Velocity would exceed Weekly Deposit Limit if transaction is accepted", account.WeeklyDepositVelocity, account.WeeklyDepositLimit, t.Amount)
		return false
	}
	if account.DailyDepositVelocity+t.Amount > account.DailyDepositLimit {
		log.Println("Daily Deposit Velocity would exceed Daily Deposit Limit if transaction is accepted", account.DailyDepositVelocity, account.DailyDepositLimit, t.Amount)
		return false
	}
	if account.DailyDepositCount+1 > account.DailyDepositCountLimit {
		log.Println("Daily Deposit Count would exceed Daily Deposit Count Limit", account.DailyDepositCount, account.DailyDepositCountLimit)
		return false
	}

	account.WeeklyDepositVelocity += t.Amount
	account.DailyDepositVelocity += t.Amount
	account.DailyDepositCount++
	account.Balance += t.Amount

	account.LastDeposit = t
	return true
}

var customers map[string]*Account

func main() {

	var defaultDailyLimit float64 = 5000
	var defaultWeeklyLimit float64 = 20000
	var defaultDailyLoadLimit = 3

	customers = map[string]*Account{}

	inputFile, err := os.Open("input.txt")

	if err != nil {
		panic(err)
	}

	testFile, err := os.Open("output.txt")

	if err != nil {
		panic(err)
	}

	inputReader := bufio.NewReader(inputFile)
	testReader := bufio.NewReader(testFile)

	transactionIDs := []string{}

	for {

		inputBytes, inputEnd := inputReader.ReadBytes('\n')

		if inputEnd != nil {
			log.Println("input end is done")
			return
		}

		t := Transaction{}

		err := json.Unmarshal(inputBytes, &t)

		if err != nil {
			panic(err)
		}

		duplicate := false
		for _, tID := range transactionIDs {
			if tID == t.ID {
				duplicate = true
				break
			}
		}

		if duplicate {
			log.Println("Duplicate transaction found; skipping", t.ID)
			continue
		}

		t.Amount, err = strconv.ParseFloat(strings.Replace(t.RawAmount, "$", "", 1), 64)

		if err != nil {
			panic(fmt.Sprint("Cannot parse float", t.RawAmount, t, err))
		}

		testBytes, testEnd := testReader.ReadBytes('\n')

		if testEnd != nil {
			panic("tests ended before input")
		}

		test := Test{}

		err = json.Unmarshal(testBytes, &test)

		if err != nil {
			panic(err)
		}

		if _, ok := customers[t.CustomerID]; !ok {
			customers[t.CustomerID] = &Account{
				CustomerID:             t.CustomerID,
				DailyDepositLimit:      defaultDailyLimit,
				WeeklyDepositLimit:     defaultWeeklyLimit,
				DailyDepositCountLimit: defaultDailyLoadLimit,
			}
		}

		account := customers[t.CustomerID]
		actual := account.Deposit(t)

		if account.CustomerID != test.CustomerID {
			fmt.Println("account and test customer ids out of sync", test.ID, account.CustomerID, test.CustomerID)
			return
		}

		if actual != test.Accepted {
			fmt.Println("account and test results do not match", actual, test.Accepted, test.ID)
			fmt.Println("Customer ID", account.CustomerID)
			fmt.Println("Balance", account.Balance)
			fmt.Println("Deposit Count", account.DailyDepositCount)
			fmt.Println("Daily Velocity", account.DailyDepositVelocity)
			fmt.Println("Weekly Velocity", account.WeeklyDepositVelocity)
			fmt.Println("Last Deposit", account.LastDeposit)
			return
		}

		transactionIDs = append(transactionIDs, t.ID)
	}
}
