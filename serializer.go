package hivego

import (
	"bytes"
	"encoding/binary"
	"time"
)

func opIdB(opName string) byte {
	id := getHiveOpId(opName)
	return byte(id)
}

func refBlockNumB(refBlockNumber uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, refBlockNumber)
	return buf
}

func refBlockPrefixB(refBlockPrefix uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, refBlockPrefix)
	return buf
}

func expTimeB(expTime string) ([]byte, error) {
	exp, err := time.Parse("2006-01-02T15:04:05", expTime)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(exp.Unix()))
	return buf, nil
}

func countOpsB(ops []HiveOperation) []byte {
	b := make([]byte, 5)
	l := binary.PutUvarint(b, uint64(len(ops)))
	return b[0:l]
}

func extensionsB() byte {
	return byte(0x00)
}

func appendVString(s string, b *bytes.Buffer) *bytes.Buffer {
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(s)))
	b.Write(vBuf[0:vLen])

	b.WriteString(s)
	return b
}

func appendVStringArray(a []string, b *bytes.Buffer) *bytes.Buffer {
	b.Write([]byte{byte(len(a))})
	for _, s := range a {
		appendVString(s, b)
	}
	return b
}

// appendAssetBytes serializes a Hive Asset in the binary wire format:
// 8-byte LE int64 amount, 1-byte precision, 7-byte null-padded symbol.
func appendAssetBytes(a Asset, b *bytes.Buffer) {
	amountBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBuf, uint64(a.Amount))
	b.Write(amountBuf)
	b.WriteByte(a.Precision)
	symbolBytes := make([]byte, 7)
	copy(symbolBytes, a.Symbol)
	b.Write(symbolBytes)
}

func serializeTx(tx Transaction) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(refBlockNumB(tx.RefBlockNum))
	buf.Write(refBlockPrefixB(tx.RefBlockPrefix))
	expTime, err := expTimeB(tx.Expiration)
	if err != nil {
		return nil, err
	}
	buf.Write(expTime)

	opsB, err := serializeOps(tx.Operations)
	if err != nil {
		return nil, err
	}
	buf.Write(opsB)
	buf.Write([]byte{extensionsB()})
	return buf.Bytes(), nil
}

func serializeOps(ops []HiveOperation) ([]byte, error) {
	var opsBuf bytes.Buffer
	opsBuf.Write(countOpsB(ops))
	for _, op := range ops {
		b, err := op.SerializeOp()
		if err != nil {
			return nil, err
		}
		opsBuf.Write(b)
	}
	return opsBuf.Bytes(), nil
}

func (o VoteOperation) SerializeOp() ([]byte, error) {
	var voteBuf bytes.Buffer
	voteBuf.Write([]byte{opIdB("vote")})
	appendVString(o.Voter, &voteBuf)
	appendVString(o.Author, &voteBuf)
	appendVString(o.Permlink, &voteBuf)

	weightBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(weightBuf, uint16(o.Weight))
	voteBuf.Write(weightBuf)

	return voteBuf.Bytes(), nil
}

func (o CustomJsonOperation) SerializeOp() ([]byte, error) {
	var jBuf bytes.Buffer
	jBuf.Write([]byte{opIdB("custom_json")})
	appendVStringArray(o.RequiredAuths, &jBuf)
	appendVStringArray(o.RequiredPostingAuths, &jBuf)
	appendVString(o.Id, &jBuf)
	appendVString(o.Json, &jBuf)

	return jBuf.Bytes(), nil
}

func (o TransferOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{opIdB("transfer")})
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendAssetBytes(o.Amount, &buf)
	appendVString(o.Memo, &buf)

	return buf.Bytes(), nil
}

func (o ClaimRewardBalanceOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{opIdB("claim_reward_balance")})
	appendVString(o.Account, &buf)
	appendAssetBytes(o.RewardHive, &buf)
	appendAssetBytes(o.RewardHbd, &buf)
	appendAssetBytes(o.RewardVests, &buf)

	return buf.Bytes(), nil
}

func (o CommentOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("comment"))
	appendVString(o.ParentAuthor, &buf)
	appendVString(o.ParentPermlink, &buf)
	appendVString(o.Author, &buf)
	appendVString(o.Permlink, &buf)
	appendVString(o.Title, &buf)
	appendVString(o.Body, &buf)
	appendVString(o.JsonMetadata, &buf)

	return buf.Bytes(), nil
}

func (o CommentOptionsOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("comment_options"))
	appendVString(o.Author, &buf)
	appendVString(o.Permlink, &buf)
	appendAssetBytes(o.MaxAcceptedPayout, &buf)

	pctBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(pctBuf, o.PercentHbd)
	buf.Write(pctBuf)

	if o.AllowVotes {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	if o.AllowCurationRewards {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	// Extensions: encode beneficiaries if present
	if len(o.Beneficiaries) > 0 {
		buf.WriteByte(1) // 1 extension follows
		buf.WriteByte(0) // type 0: comment_payout_beneficiaries
		vBuf := make([]byte, 5)
		vLen := binary.PutUvarint(vBuf, uint64(len(o.Beneficiaries)))
		buf.Write(vBuf[:vLen])
		for _, b := range o.Beneficiaries {
			appendVString(b.Account, &buf)
			wBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(wBuf, b.Weight)
			buf.Write(wBuf)
		}
	} else {
		buf.WriteByte(0) // no extensions
	}

	return buf.Bytes(), nil
}

func (o DeleteCommentOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("delete_comment"))
	appendVString(o.Author, &buf)
	appendVString(o.Permlink, &buf)

	return buf.Bytes(), nil
}

func (o TransferToVestingOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("transfer_to_vesting"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendAssetBytes(o.Amount, &buf)

	return buf.Bytes(), nil
}

func (o WithdrawVestingOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("withdraw_vesting"))
	appendVString(o.Account, &buf)
	appendAssetBytes(o.VestingShares, &buf)

	return buf.Bytes(), nil
}

func (o DelegateVestingSharesOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("delegate_vesting_shares"))
	appendVString(o.Delegator, &buf)
	appendVString(o.Delegatee, &buf)
	appendAssetBytes(o.VestingShares, &buf)

	return buf.Bytes(), nil
}

func (o AccountWitnessVoteOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("account_witness_vote"))
	appendVString(o.Account, &buf)
	appendVString(o.Witness, &buf)
	if o.Approve {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	return buf.Bytes(), nil
}

func (o TransferToSavingsOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("transfer_to_savings"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendAssetBytes(o.Amount, &buf)
	appendVString(o.Memo, &buf)

	return buf.Bytes(), nil
}

func (o TransferFromSavingsOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("transfer_from_savings"))
	appendVString(o.From, &buf)
	reqIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(reqIdBuf, o.RequestId)
	buf.Write(reqIdBuf)
	appendVString(o.To, &buf)
	appendAssetBytes(o.Amount, &buf)
	appendVString(o.Memo, &buf)

	return buf.Bytes(), nil
}
