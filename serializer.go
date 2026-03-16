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
//
// condenser_api uses legacy (pre-HF26) serialization, which requires the original
// Steem symbol names on the wire: HIVE→"STEEM", HBD→"SBD". All other symbols
// (e.g. "VESTS") are written as-is. See HIVE_ASSET_NUM / OBSOLETE_SYMBOL_SER in
// hive/libraries/protocol/include/hive/protocol/asset_symbol.hpp.
func appendAssetBytes(a Asset, b *bytes.Buffer) {
	amountBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBuf, uint64(a.Amount))
	b.Write(amountBuf)
	b.WriteByte(a.Precision)
	wireSymbol := a.Symbol
	switch a.Symbol {
	case "HIVE":
		wireSymbol = "STEEM"
	case "HBD":
		wireSymbol = "SBD"
	}
	symbolBytes := make([]byte, 7)
	copy(symbolBytes, wireSymbol)
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

func (o SetWithdrawVestingRouteOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("set_withdraw_vesting_route"))
	appendVString(o.FromAccount, &buf)
	appendVString(o.ToAccount, &buf)
	pctBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(pctBuf, o.Percent)
	buf.Write(pctBuf)
	if o.AutoVest {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	return buf.Bytes(), nil
}

func (o CancelTransferFromSavingsOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("cancel_transfer_from_savings"))
	appendVString(o.From, &buf)
	reqIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(reqIdBuf, o.RequestId)
	buf.Write(reqIdBuf)
	return buf.Bytes(), nil
}

func (o RecurrentTransferOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("recurrent_transfer"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendAssetBytes(o.Amount, &buf)
	appendVString(o.Memo, &buf)
	recBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(recBuf, o.Recurrence)
	buf.Write(recBuf)
	execBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(execBuf, o.Executions)
	buf.Write(execBuf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o AccountUpdate2Operation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("account_update2"))
	appendVString(o.Account, &buf)
	if o.Owner == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Owner, &buf)
	}
	if o.Active == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Active, &buf)
	}
	if o.Posting == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Posting, &buf)
	}
	if o.MemoKey == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		buf.Write(o.MemoKey.SerializeCompressed())
	}
	appendVString(o.JsonMetadata, &buf)
	appendVString(o.PostingJsonMetadata, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o ConvertOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("convert"))
	appendVString(o.Owner, &buf)
	reqIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(reqIdBuf, o.RequestId)
	buf.Write(reqIdBuf)
	appendAssetBytes(o.Amount, &buf)
	return buf.Bytes(), nil
}

func (o CollateralizedConvertOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("collateralized_convert"))
	appendVString(o.Owner, &buf)
	reqIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(reqIdBuf, o.RequestId)
	buf.Write(reqIdBuf)
	appendAssetBytes(o.Amount, &buf)
	return buf.Bytes(), nil
}

func appendAuthorityBytes(a *Authority, b *bytes.Buffer) {
	threshBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(threshBuf, a.WeightThreshold)
	b.Write(threshBuf)

	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(a.AccountAuths)))
	b.Write(vBuf[:vLen])
	for _, aa := range a.AccountAuths {
		appendVString(aa.Account, b)
		wBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(wBuf, aa.Weight)
		b.Write(wBuf)
	}

	vLen = binary.PutUvarint(vBuf, uint64(len(a.KeyAuths)))
	b.Write(vBuf[:vLen])
	for _, ka := range a.KeyAuths {
		b.Write(ka.Key.SerializeCompressed())
		wBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(wBuf, ka.Weight)
		b.Write(wBuf)
	}
}

func appendInt64Array(ids []int64, b *bytes.Buffer) {
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(ids)))
	b.Write(vBuf[:vLen])
	for _, id := range ids {
		idBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(idBuf, uint64(id))
		b.Write(idBuf)
	}
}

func (o LimitOrderCreateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("limit_order_create"))
	appendVString(o.Owner, &buf)
	orderIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(orderIdBuf, o.OrderId)
	buf.Write(orderIdBuf)
	appendAssetBytes(o.AmountToSell, &buf)
	appendAssetBytes(o.MinToReceive, &buf)
	if o.FillOrKill {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	expB, err := expTimeB(o.Expiration)
	if err != nil {
		return nil, err
	}
	buf.Write(expB)
	return buf.Bytes(), nil
}

func (o LimitOrderCancelOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("limit_order_cancel"))
	appendVString(o.Owner, &buf)
	orderIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(orderIdBuf, o.OrderId)
	buf.Write(orderIdBuf)
	return buf.Bytes(), nil
}

func (o AccountWitnessProxyOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("account_witness_proxy"))
	appendVString(o.Account, &buf)
	appendVString(o.Proxy, &buf)
	return buf.Bytes(), nil
}

func (o CreateProposalOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("create_proposal"))
	appendVString(o.Creator, &buf)
	appendVString(o.Receiver, &buf)
	startB, err := expTimeB(o.StartDate)
	if err != nil {
		return nil, err
	}
	buf.Write(startB)
	endB, err := expTimeB(o.EndDate)
	if err != nil {
		return nil, err
	}
	buf.Write(endB)
	appendAssetBytes(o.DailyPay, &buf)
	appendVString(o.Subject, &buf)
	appendVString(o.Permlink, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o UpdateProposalVotesOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("update_proposal_votes"))
	appendVString(o.Voter, &buf)
	appendInt64Array(o.ProposalIds, &buf)
	if o.Approve {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o RemoveProposalOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("remove_proposal"))
	appendVString(o.ProposalOwner, &buf)
	appendInt64Array(o.ProposalIds, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o UpdateProposalOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("update_proposal"))
	idBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBuf, uint64(o.ProposalId))
	buf.Write(idBuf)
	appendVString(o.Creator, &buf)
	appendAssetBytes(o.DailyPay, &buf)
	appendVString(o.Subject, &buf)
	appendVString(o.Permlink, &buf)
	// EndDate is encoded in the extensions array as a static_variant.
	// update_proposal_end_date has type_id=1 (second type in the variant after void_t).
	if o.EndDate == "" {
		buf.WriteByte(0) // empty extensions array
	} else {
		buf.WriteByte(1) // 1 extension
		buf.WriteByte(1) // static_variant type_id=1 (update_proposal_end_date)
		endB, err := expTimeB(o.EndDate)
		if err != nil {
			return nil, err
		}
		buf.Write(endB)
	}
	return buf.Bytes(), nil
}

func (o ClaimAccountOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("claim_account"))
	appendVString(o.Creator, &buf)
	appendAssetBytes(o.Fee, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o FeedPublishOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("feed_publish"))
	appendVString(o.Publisher, &buf)
	appendAssetBytes(o.ExchangeRate.Base, &buf)
	appendAssetBytes(o.ExchangeRate.Quote, &buf)
	return buf.Bytes(), nil
}

func appendChainPropertiesBytes(p ChainProperties, b *bytes.Buffer) {
	appendAssetBytes(p.AccountCreationFee, b)
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, p.MaximumBlockSize)
	b.Write(sizeBuf)
	rateBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(rateBuf, p.HbdInterestRate)
	b.Write(rateBuf)
}

func appendVBytes(data []byte, b *bytes.Buffer) {
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(data)))
	b.Write(vBuf[:vLen])
	b.Write(data)
}

func (o AccountCreateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("account_create"))
	appendAssetBytes(o.Fee, &buf)
	appendVString(o.Creator, &buf)
	appendVString(o.NewAccountName, &buf)
	appendAuthorityBytes(&o.Owner, &buf)
	appendAuthorityBytes(&o.Active, &buf)
	appendAuthorityBytes(&o.Posting, &buf)
	if o.MemoKey != nil {
		buf.Write(o.MemoKey.SerializeCompressed())
	} else {
		buf.Write(make([]byte, 33))
	}
	appendVString(o.JsonMetadata, &buf)
	return buf.Bytes(), nil
}

func (o CreateClaimedAccountOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("create_claimed_account"))
	appendVString(o.Creator, &buf)
	appendVString(o.NewAccountName, &buf)
	appendAuthorityBytes(&o.Owner, &buf)
	appendAuthorityBytes(&o.Active, &buf)
	appendAuthorityBytes(&o.Posting, &buf)
	if o.MemoKey != nil {
		buf.Write(o.MemoKey.SerializeCompressed())
	} else {
		buf.Write(make([]byte, 33))
	}
	appendVString(o.JsonMetadata, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o AccountUpdateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("account_update"))
	appendVString(o.Account, &buf)
	if o.Owner == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Owner, &buf)
	}
	if o.Active == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Active, &buf)
	}
	if o.Posting == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		appendAuthorityBytes(o.Posting, &buf)
	}
	if o.MemoKey != nil {
		buf.Write(o.MemoKey.SerializeCompressed())
	} else {
		buf.Write(make([]byte, 33))
	}
	appendVString(o.JsonMetadata, &buf)
	return buf.Bytes(), nil
}

func (o WitnessUpdateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("witness_update"))
	appendVString(o.Owner, &buf)
	appendVString(o.Url, &buf)
	if o.BlockSigningKey != nil {
		buf.Write(o.BlockSigningKey.SerializeCompressed())
	} else {
		buf.Write(make([]byte, 33))
	}
	appendChainPropertiesBytes(o.Props, &buf)
	appendAssetBytes(o.Fee, &buf)
	return buf.Bytes(), nil
}

func (o WitnessSetPropertiesOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("witness_set_properties"))
	appendVString(o.Owner, &buf)
	// flat_map: varint count, then sorted key-value pairs
	keys := make([]string, 0, len(o.Props))
	for k := range o.Props {
		keys = append(keys, k)
	}
	// sort inline to avoid import cycle (sort already imported in hive_ops.go but not here)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(keys)))
	buf.Write(vBuf[:vLen])
	for _, k := range keys {
		appendVString(k, &buf)
		appendVBytes(o.Props[k], &buf)
	}
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o RequestAccountRecoveryOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("request_account_recovery"))
	appendVString(o.RecoveryAccount, &buf)
	appendVString(o.AccountToRecover, &buf)
	appendAuthorityBytes(&o.NewOwnerAuthority, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o RecoverAccountOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("recover_account"))
	appendVString(o.AccountToRecover, &buf)
	appendAuthorityBytes(&o.NewOwnerAuthority, &buf)
	appendAuthorityBytes(&o.RecentOwnerAuthority, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o ChangeRecoveryAccountOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("change_recovery_account"))
	appendVString(o.AccountToRecover, &buf)
	appendVString(o.NewRecoveryAccount, &buf)
	buf.WriteByte(0) // no extensions
	return buf.Bytes(), nil
}

func (o EscrowTransferOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("escrow_transfer"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendVString(o.Agent, &buf)
	idBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(idBuf, o.EscrowId)
	buf.Write(idBuf)
	appendAssetBytes(o.HbdAmount, &buf)
	appendAssetBytes(o.HiveAmount, &buf)
	appendAssetBytes(o.Fee, &buf)
	ratB, err := expTimeB(o.RatificationDeadline)
	if err != nil {
		return nil, err
	}
	buf.Write(ratB)
	expB, err := expTimeB(o.EscrowExpiration)
	if err != nil {
		return nil, err
	}
	buf.Write(expB)
	appendVString(o.JsonMeta, &buf)
	return buf.Bytes(), nil
}

func (o EscrowApproveOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("escrow_approve"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendVString(o.Agent, &buf)
	appendVString(o.Who, &buf)
	idBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(idBuf, o.EscrowId)
	buf.Write(idBuf)
	if o.Approve {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	return buf.Bytes(), nil
}

func (o EscrowDisputeOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("escrow_dispute"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendVString(o.Agent, &buf)
	appendVString(o.Who, &buf)
	idBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(idBuf, o.EscrowId)
	buf.Write(idBuf)
	return buf.Bytes(), nil
}

func (o EscrowReleaseOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("escrow_release"))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	appendVString(o.Agent, &buf)
	appendVString(o.Who, &buf)
	appendVString(o.Receiver, &buf)
	idBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(idBuf, o.EscrowId)
	buf.Write(idBuf)
	appendAssetBytes(o.HbdAmount, &buf)
	appendAssetBytes(o.HiveAmount, &buf)
	return buf.Bytes(), nil
}

func (o CustomOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("custom"))
	appendVStringArray(o.RequiredAuths, &buf)
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, o.Id)
	buf.Write(b)
	appendVString(o.Data, &buf)
	return buf.Bytes(), nil
}

func (o CustomBinaryOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB("custom_binary"))
	appendVStringArray(o.RequiredOwnerAuths, &buf)
	appendVStringArray(o.RequiredActiveAuths, &buf)
	appendVStringArray(o.RequiredPostingAuths, &buf)
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(o.RequiredAuths)))
	buf.Write(vBuf[:vLen])
	for i := range o.RequiredAuths {
		appendAuthorityBytes(&o.RequiredAuths[i], &buf)
	}
	appendVString(o.Id, &buf)
	appendVBytes(o.Data, &buf)
	return buf.Bytes(), nil
}

func (o DeclineVotingRightsOperation) SerializeOp() ([]byte, error) {
	if o.Decline && !o.IUnderstandThisIsIrreversible {
		return nil, ErrDeclineVotingRightsNotConfirmed
	}
	var buf bytes.Buffer
	buf.WriteByte(opIdB("decline_voting_rights"))
	appendVString(o.Account, &buf)
	if o.Decline {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	return buf.Bytes(), nil
}
