package hivego

import "time"

func getTestVoteOp() HiveOperation {
	return VoteOperation{
		Voter:    "xeroc",
		Author:   "xeroc",
		Permlink: "piston",
		Weight:   10000,
	}
}

func getTestCustomJsonOp() HiveOperation {
	return CustomJsonOperation{
		RequiredAuths:        []string{},
		RequiredPostingAuths: []string{"xeroc"},
		Id:                   "test-id",
		Json:                 "{\"testk\":\"testv\"}",
	}
}

func getTwoTestOps() []HiveOperation {
	return []HiveOperation{getTestVoteOp(), getTestCustomJsonOp()}
}

func getTestTx(ops []HiveOperation) Transaction {
	exp, _ := time.Parse("2006-01-02T15:04:05", "2016-08-08T12:24:17")
	expStr := exp.Format("2006-01-02T15:04:05")

	return Transaction{
		RefBlockNum:    36029,
		RefBlockPrefix: 1164960351,
		Expiration:     expStr,
		Operations:     ops,
	}
}

func getTestVoteTx() Transaction {
	return getTestTx([]HiveOperation{getTestVoteOp()})
}
