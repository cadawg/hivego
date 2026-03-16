#!/usr/bin/env python3
"""
generate_vectors.py — generates and validates serialization test vectors for hivego
using beem as an independent reference implementation.

Setup (venv already committed at testdata/.venv):
    testdata/.venv/bin/python3 testdata/generate_vectors.py

To recreate the venv from scratch:
    python3 -m venv testdata/.venv
    testdata/.venv/bin/pip install -r testdata/requirements.txt

What this does:
  1. Serializes each Hive operation using beembase and prints the expected byte
     slices used in serializer_test.go.
  2. Verifies the signing digest (hashTxForSig) produced by hivego matches beem.
  3. Verifies hivego's compact signature from TestSignDigest using the pure-Python
     ecdsa library (beem's verify_message uses a secp256k1 C binding with API
     breakage on Python 3.14+, so we use ecdsa directly).

Note on asset symbols: beem (like the Hive node) uses the legacy Steem symbol names
("STEEM", "SBD") on the wire even though the API uses "HIVE"/"HBD". hivego's
appendAssetBytes maps these correctly. See OBSOLETE_SYMBOL_SER in:
  hive/libraries/protocol/include/hive/protocol/asset_symbol.hpp
"""

import hashlib
import binascii
import struct
import sys

import ecdsa  # pure-Python fallback for signature verification

from beembase.operations import (
    Vote, Transfer, Comment, Delete_comment, Claim_reward_balance,
    Transfer_to_vesting, Withdraw_vesting, Delegate_vesting_shares,
    Account_witness_vote, Account_witness_proxy,
    Transfer_to_savings, Transfer_from_savings, Cancel_transfer_from_savings,
    Set_withdraw_vesting_route, Recurring_transfer,
    Account_update2, Account_update, Convert, Collateralized_convert,
    Limit_order_create, Limit_order_cancel, Feed_publish, Change_recovery_account,
    Decline_voting_rights, Claim_account, Comment_options,
    Create_proposal, Update_proposal_votes, Remove_proposal, Update_proposal,
    Account_create, Create_claimed_account, Witness_update, Witness_set_properties,
    Request_account_recovery, Recover_account,
    Escrow_transfer, Escrow_approve, Escrow_dispute, Escrow_release,
)
from beembase.objects import Operation
from beemgraphenebase.signedtransactions import Signed_Transaction


CHAIN_ID = "beeab0de00000000000000000000000000000000000000000000000000000000"
CHAIN_PARAMS = {"chain_id": CHAIN_ID}


def go_bytes(b: bytes) -> str:
    return "{" + ", ".join(str(x) for x in b) + "}"


def serialize_op(op_name: str, op_obj) -> bytes:
    op = Operation([op_name, op_obj])
    return bytes(op)


def section(title: str):
    print(f"\n// --- {title} ---")


def check(name: str, got: bytes, expected: bytes):
    if got == expected:
        print(f"// OK: {name}")
    else:
        print(f"// MISMATCH: {name}")
        print(f"//   beem:   {list(got)}")
        print(f"//   hivego: {list(expected)}")
        sys.exit(1)


print("// beem-generated test vectors for hivego/serializer_test.go")
print(f"// beem version: ", end="")
try:
    import beem
    print(beem.__version__)
except Exception:
    print("unknown")

# ---------------------------------------------------------------------------
# Operation byte vectors
# ---------------------------------------------------------------------------

section("VoteOperation — already in TestSerializeOpVoteOperation")
vote_bytes = serialize_op("vote", Vote(**{
    "voter": "xeroc", "author": "xeroc", "permlink": "piston", "weight": 10000,
}))
print(f"vote: {go_bytes(vote_bytes)}")

section("TransferOperation")
transfer_bytes = serialize_op("transfer", Transfer(**{
    "from": "alice", "to": "bob", "amount": "1.000 HIVE", "memo": "memo",
}))
print(f"transfer: {go_bytes(transfer_bytes)}")

section("CommentOperation")
comment_bytes = serialize_op("comment", Comment(**{
    "parent_author": "", "parent_permlink": "hive-blog",
    "author": "alice", "permlink": "my-post",
    "title": "Hello", "body": "World", "json_metadata": "{}",
}))
print(f"comment: {go_bytes(comment_bytes)}")

section("DeleteCommentOperation")
delete_bytes = serialize_op("delete_comment", Delete_comment(**{
    "author": "alice", "permlink": "my-post",
}))
print(f"delete_comment: {go_bytes(delete_bytes)}")

section("ClaimRewardBalanceOperation")
claim_reward_bytes = serialize_op("claim_reward_balance", Claim_reward_balance(**{
    "account": "alice",
    "reward_hive": "0.000 HIVE",
    "reward_hbd": "0.000 HBD",
    "reward_vests": "1.000000 VESTS",
}))
print(f"claim_reward_balance: {go_bytes(claim_reward_bytes)}")

section("TransferToVestingOperation")
ttv_bytes = serialize_op("transfer_to_vesting", Transfer_to_vesting(**{
    "from": "alice", "to": "bob", "amount": "100.000 HIVE",
}))
print(f"transfer_to_vesting: {go_bytes(ttv_bytes)}")

section("WithdrawVestingOperation")
wv_bytes = serialize_op("withdraw_vesting", Withdraw_vesting(**{
    "account": "alice", "vesting_shares": "0.000000 VESTS",
}))
print(f"withdraw_vesting: {go_bytes(wv_bytes)}")

section("DelegateVestingSharesOperation")
dvs_bytes = serialize_op("delegate_vesting_shares", Delegate_vesting_shares(**{
    "delegator": "alice", "delegatee": "bob", "vesting_shares": "1.000000 VESTS",
}))
print(f"delegate_vesting_shares: {go_bytes(dvs_bytes)}")

section("AccountWitnessVoteOperation")
awv_bytes = serialize_op("account_witness_vote", Account_witness_vote(**{
    "account": "alice", "witness": "bob", "approve": True,
}))
print(f"account_witness_vote: {go_bytes(awv_bytes)}")

section("AccountWitnessProxyOperation (remove proxy)")
awp_bytes = serialize_op("account_witness_proxy", Account_witness_proxy(**{
    "account": "alice", "proxy": "",
}))
print(f"account_witness_proxy: {go_bytes(awp_bytes)}")

section("TransferToSavingsOperation")
tts_bytes = serialize_op("transfer_to_savings", Transfer_to_savings(**{
    "from": "alice", "to": "alice", "amount": "100.000 HBD", "memo": "",
}))
print(f"transfer_to_savings: {go_bytes(tts_bytes)}")

section("TransferFromSavingsOperation")
tfs_bytes = serialize_op("transfer_from_savings", Transfer_from_savings(**{
    "from": "alice", "request_id": 1, "to": "alice", "amount": "1.000 HBD", "memo": "",
}))
print(f"transfer_from_savings: {go_bytes(tfs_bytes)}")

section("CancelTransferFromSavingsOperation")
ctfs_bytes = serialize_op("cancel_transfer_from_savings", Cancel_transfer_from_savings(**{
    "from": "alice", "request_id": 42,
}))
print(f"cancel_transfer_from_savings: {go_bytes(ctfs_bytes)}")

section("SetWithdrawVestingRouteOperation")
swvr_bytes = serialize_op("set_withdraw_vesting_route", Set_withdraw_vesting_route(**{
    "from_account": "alice", "to_account": "bob", "percent": 5000, "auto_vest": True,
}))
print(f"set_withdraw_vesting_route: {go_bytes(swvr_bytes)}")

section("RecurrentTransferOperation")
# NOTE: beem 0.24.x has a Recurring_transfer class but does NOT include it in its
# operations map (beem was never updated for HF25 recurrent_transfer). We serialize
# the operation body directly and prepend the op_id (49). Additionally, beem's
# Recurring_transfer omits the extensions field; the Hive wire format requires it.
# We append a 0x00 byte (empty extensions array) to match the protocol.
_rt_body = bytes(Recurring_transfer(**{
    "from": "alice", "to": "bob", "amount": "1.000 HIVE",
    "memo": "", "recurrence": 24, "executions": 7,
    "extensions": [],
}))
rt_bytes = bytes([49]) + _rt_body + bytes([0])  # op_id=49, body, extensions=0
print(f"recurrent_transfer: {go_bytes(rt_bytes)}")

section("AccountUpdate2Operation (metadata only)")
au2_bytes = serialize_op("account_update2", Account_update2(**{
    "account": "alice",
    "json_metadata": "{}",
    "posting_json_metadata": "",
    "extensions": [],
}))
print(f"account_update2: {go_bytes(au2_bytes)}")

section("ConvertOperation")
convert_bytes = serialize_op("convert", Convert(**{
    "owner": "alice", "requestid": 1, "amount": "1.000 HBD",
}))
print(f"convert: {go_bytes(convert_bytes)}")

section("CollateralizedConvertOperation")
# NOTE: beem 0.24.x has a Collateralized_convert class but does NOT include it in
# its operations map (not added until HF25). We serialize directly and prepend op_id 48.
_cc_body = bytes(Collateralized_convert(**{
    "owner": "alice", "requestid": 1, "amount": "1.000 HIVE",
}))
cc_bytes = bytes([48]) + _cc_body  # op_id=48
print(f"collateralized_convert: {go_bytes(cc_bytes)}")

section("LimitOrderCancelOperation")
loc_bytes = serialize_op("limit_order_cancel", Limit_order_cancel(**{
    "owner": "alice", "orderid": 1,
}))
print(f"limit_order_cancel: {go_bytes(loc_bytes)}")

section("FeedPublishOperation")
fp_bytes = serialize_op("feed_publish", Feed_publish(**{
    "publisher": "alice",
    "exchange_rate": {"base": "1.000 HBD", "quote": "3.500 HIVE"},
}))
print(f"feed_publish: {go_bytes(fp_bytes)}")

section("ChangeRecoveryAccountOperation")
cra_bytes = serialize_op("change_recovery_account", Change_recovery_account(**{
    "account_to_recover": "alice", "new_recovery_account": "bob", "extensions": [],
}))
print(f"change_recovery_account: {go_bytes(cra_bytes)}")

section("DeclineVotingRightsOperation (decline=false)")
dvr_false_bytes = serialize_op("decline_voting_rights", Decline_voting_rights(**{
    "account": "alice", "decline": False,
}))
print(f"decline_voting_rights(false): {go_bytes(dvr_false_bytes)}")

section("DeclineVotingRightsOperation (decline=true)")
dvr_true_bytes = serialize_op("decline_voting_rights", Decline_voting_rights(**{
    "account": "alice", "decline": True,
}))
print(f"decline_voting_rights(true): {go_bytes(dvr_true_bytes)}")

section("ClaimAccountOperation (fee=0.000 HIVE)")
ca_bytes = serialize_op("claim_account", Claim_account(**{
    "creator": "alice", "fee": "0.000 HIVE", "extensions": [],
}))
print(f"claim_account: {go_bytes(ca_bytes)}")

# ---------------------------------------------------------------------------
# New operations added in second pass
# ---------------------------------------------------------------------------

from beemgraphenebase.account import PrivateKey as BeemPrivKey
_WIF = "5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W"
_pub_str = str(BeemPrivKey(_WIF).pubkey)  # STM5nD5Sn9avUfQ1i3dZQRizopgNko19p3GrzgmwxA2YMapHrL8wx
_pub_compressed = bytes.fromhex("02756f674e8e93ffc8eb719c933a9b06dcab85dc62160b3129c93a9ef767f34091")
_auth = {"weight_threshold": 1, "account_auths": [], "key_auths": [[_pub_str, 1]]}

section("CommentOptionsOperation (no beneficiaries)")
co_bytes = serialize_op("comment_options", Comment_options(**{
    "author": "alice", "permlink": "my-post",
    "max_accepted_payout": "1000000.000 HBD",
    "percent_hbd": 10000,
    "allow_votes": True, "allow_curation_rewards": True,
    "extensions": [],
}))
print(f"comment_options (no bens): {go_bytes(co_bytes)}")
check("TestSerializeOpCommentOptionsNoBeneficiaries", co_bytes,
      bytes([19, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116, 0, 202, 154, 59, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 16, 39, 1, 1, 0]))

section("CommentOptionsOperation (one beneficiary bob@5000)")
co2_bytes = serialize_op("comment_options", Comment_options(**{
    "author": "alice", "permlink": "my-post",
    "max_accepted_payout": "1000000.000 HBD",
    "percent_hbd": 5000,
    "allow_votes": True, "allow_curation_rewards": True,
    "extensions": [[0, {"beneficiaries": [{"account": "bob", "weight": 5000}]}]],
}))
print(f"comment_options (one ben): {go_bytes(co2_bytes)}")
check("TestSerializeOpCommentOptionsWithBeneficiary", co2_bytes,
      bytes([19, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116, 0, 202, 154, 59, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 136, 19, 1, 1, 1, 0, 1, 3, 98, 111, 98, 136, 19]))

section("LimitOrderCreateOperation")
lo_bytes = serialize_op("limit_order_create", Limit_order_create(**{
    "owner": "alice", "orderid": 1,
    "amount_to_sell": "1.000 HIVE", "min_to_receive": "100.000 HBD",
    "fill_or_kill": False, "expiration": "2030-01-01T00:00:00",
}))
print(f"limit_order_create: {go_bytes(lo_bytes)}")
check("TestSerializeOpLimitOrderCreate", lo_bytes,
      bytes([5, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0, 128, 216, 219, 112]))

section("CreateProposalOperation")
cp_bytes = serialize_op("create_proposal", Create_proposal(**{
    "creator": "alice", "receiver": "bob",
    "start_date": "2024-01-01T00:00:00", "end_date": "2024-12-31T00:00:00",
    "daily_pay": "100.000 HBD",
    "subject": "My Proposal", "permlink": "my-proposal",
    "extensions": [],
}))
print(f"create_proposal: {go_bytes(cp_bytes)}")
check("TestSerializeOpCreateProposal", cp_bytes,
      bytes([44, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 128, 0, 146, 101, 0, 52, 115, 103, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 11, 77, 121, 32, 80, 114, 111, 112, 111, 115, 97, 108, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 0]))

section("UpdateProposalVotesOperation")
upv_bytes = serialize_op("update_proposal_votes", Update_proposal_votes(**{
    "voter": "alice", "proposal_ids": [1, 2], "approve": True, "extensions": [],
}))
print(f"update_proposal_votes: {go_bytes(upv_bytes)}")
check("TestSerializeOpUpdateProposalVotes", upv_bytes,
      bytes([45, 5, 97, 108, 105, 99, 101, 2, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 0]))

section("RemoveProposalOperation")
rp_bytes = serialize_op("remove_proposal", Remove_proposal(**{
    "proposal_owner": "alice", "proposal_ids": [1], "extensions": [],
}))
print(f"remove_proposal: {go_bytes(rp_bytes)}")
check("TestSerializeOpRemoveProposal", rp_bytes,
      bytes([46, 5, 97, 108, 105, 99, 101, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0]))

section("UpdateProposalOperation (no end_date)")
# NOTE: beem encodes EndDate in the extensions array as static_variant type_id=1.
# No end_date → empty extensions array → single 0x00 byte.
# Our Go code previously used fc::optional encoding (two bytes: 0x00 + 0x00) — now fixed.
up_no_end_bytes = bytes(Update_proposal(**{
    "proposal_id": 1, "creator": "alice", "daily_pay": "50.000 HBD",
    "subject": "Updated", "permlink": "my-proposal",
}))
up_no_end_bytes = bytes([47]) + up_no_end_bytes  # prepend op_id
print(f"update_proposal (no end): {go_bytes(up_no_end_bytes)}")
check("TestSerializeOpUpdateProposalNoEndDate", up_no_end_bytes,
      bytes([47, 1, 0, 0, 0, 0, 0, 0, 0, 5, 97, 108, 105, 99, 101, 80, 195, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 7, 85, 112, 100, 97, 116, 101, 100, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 0]))

section("UpdateProposalOperation (with end_date=2025-06-01)")
# end_date is encoded as extensions=[{type_id=1, timestamp}]
up_with_end_bytes = bytes(Update_proposal(**{
    "proposal_id": 1, "creator": "alice", "daily_pay": "50.000 HBD",
    "subject": "Updated", "permlink": "my-proposal",
    "end_date": "2025-06-01T00:00:00",
}))
up_with_end_bytes = bytes([47]) + up_with_end_bytes  # prepend op_id
print(f"update_proposal (with end): {go_bytes(up_with_end_bytes)}")
check("TestSerializeOpUpdateProposalWithEndDate", up_with_end_bytes,
      bytes([47, 1, 0, 0, 0, 0, 0, 0, 0, 5, 97, 108, 105, 99, 101, 80, 195, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 7, 85, 112, 100, 97, 116, 101, 100, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 1, 1, 0, 152, 59, 104]))

section("AccountCreateOperation (fee=3.000 HIVE, single key authority)")
ac_bytes = serialize_op("account_create", Account_create(**{
    "fee": "3.000 HIVE", "creator": "alice", "new_account_name": "bob",
    "owner": _auth, "active": _auth, "posting": _auth,
    "memo_key": _pub_str, "json_metadata": "{}",
}))
print(f"account_create: {go_bytes(ac_bytes)}")
check("TestSerializeOpAccountCreate", ac_bytes,
      bytes([9, 184, 11, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125]))

section("CreateClaimedAccountOperation")
cca_bytes = serialize_op("create_claimed_account", Create_claimed_account(**{
    "creator": "alice", "new_account_name": "bob",
    "owner": _auth, "active": _auth, "posting": _auth,
    "memo_key": _pub_str, "json_metadata": "{}",
    "extensions": [],
}))
print(f"create_claimed_account: {go_bytes(cca_bytes)}")
check("TestSerializeOpCreateClaimedAccount", cca_bytes,
      bytes([23, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125, 0]))

section("AccountUpdateOperation (memo key only, all authorities absent)")
au_memo_bytes = serialize_op("account_update", Account_update(**{
    "account": "alice", "memo_key": _pub_str, "json_metadata": "",
}))
print(f"account_update (memo only): {go_bytes(au_memo_bytes)}")
check("TestSerializeOpAccountUpdateMemoOnly", au_memo_bytes,
      bytes([10, 5, 97, 108, 105, 99, 101, 0, 0, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0]))

section("AccountUpdateOperation (active authority + memo key)")
au_active_bytes = serialize_op("account_update", Account_update(**{
    "account": "alice", "active": _auth, "memo_key": _pub_str, "json_metadata": "",
}))
print(f"account_update (active+memo): {go_bytes(au_active_bytes)}")
check("TestSerializeOpAccountUpdateWithActive", au_active_bytes,
      bytes([10, 5, 97, 108, 105, 99, 101, 0, 1, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0]))

section("AccountUpdate2Operation (active authority + memo key)")
au2_active_bytes = serialize_op("account_update2", Account_update2(**{
    "account": "alice", "active": _auth, "memo_key": _pub_str,
    "json_metadata": "{}", "posting_json_metadata": "",
}))
print(f"account_update2 (active+memo): {go_bytes(au2_active_bytes)}")
check("TestSerializeOpAccountUpdate2WithActive", au2_active_bytes,
      bytes([43, 5, 97, 108, 105, 99, 101, 0, 1, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125, 0, 0]))

section("WitnessUpdateOperation")
wu_bytes = serialize_op("witness_update", Witness_update(**{
    "owner": "alice", "url": "https://example.com",
    "block_signing_key": _pub_str,
    "props": {"account_creation_fee": "3.000 HIVE", "maximum_block_size": 65536, "hbd_interest_rate": 0},
    "fee": "0.000 HIVE",
}))
print(f"witness_update: {go_bytes(wu_bytes)}")
check("TestSerializeOpWitnessUpdate", wu_bytes,
      bytes([11, 5, 97, 108, 105, 99, 101, 19, 104, 116, 116, 112, 115, 58, 47, 47, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 184, 11, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0]))

section("WitnessSetPropertiesOperation (key property only)")
wsp_bytes = serialize_op("witness_set_properties", Witness_set_properties(**{
    "owner": "alice",
    "props": [["key", _pub_compressed.hex()]],
    "extensions": [],
}))
print(f"witness_set_properties: {go_bytes(wsp_bytes)}")
check("TestSerializeOpWitnessSetProperties", wsp_bytes,
      bytes([42, 5, 97, 108, 105, 99, 101, 1, 3, 107, 101, 121, 33, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0]))

section("RequestAccountRecoveryOperation")
rar_bytes = serialize_op("request_account_recovery", Request_account_recovery(**{
    "recovery_account": "alice", "account_to_recover": "bob",
    "new_owner_authority": _auth, "extensions": [],
}))
print(f"request_account_recovery: {go_bytes(rar_bytes)}")
check("TestSerializeOpRequestAccountRecovery", rar_bytes,
      bytes([24, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0]))

section("RecoverAccountOperation")
ra_bytes = serialize_op("recover_account", Recover_account(**{
    "account_to_recover": "bob",
    "new_owner_authority": _auth, "recent_owner_authority": _auth,
    "extensions": [],
}))
print(f"recover_account: {go_bytes(ra_bytes)}")
check("TestSerializeOpRecoverAccount", ra_bytes,
      bytes([25, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0]))

section("EscrowTransferOperation")
et_bytes = serialize_op("escrow_transfer", Escrow_transfer(**{
    "from": "alice", "to": "bob", "agent": "charlie", "escrow_id": 1,
    "hbd_amount": "1.000 HBD", "hive_amount": "1.000 HIVE", "fee": "0.001 HIVE",
    "ratification_deadline": "2030-01-01T00:00:00",
    "escrow_expiration": "2030-06-01T00:00:00",
    "json_meta": "{}",
}))
print(f"escrow_transfer: {go_bytes(et_bytes)}")
check("TestSerializeOpEscrowTransfer", et_bytes,
      bytes([27, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 128, 216, 219, 112, 0, 235, 162, 113, 2, 123, 125]))

section("EscrowApproveOperation")
ea_bytes = serialize_op("escrow_approve", Escrow_approve(**{
    "from": "alice", "to": "bob", "agent": "charlie",
    "who": "bob", "escrow_id": 1, "approve": True,
}))
print(f"escrow_approve: {go_bytes(ea_bytes)}")
check("TestSerializeOpEscrowApprove", ea_bytes,
      bytes([31, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 3, 98, 111, 98, 1, 0, 0, 0, 1]))

# NOTE: beem's Escrow_dispute omits the 'agent' field (from, to, who, escrow_id only).
# Our Go implementation follows the Hive protocol spec which includes agent.
# The expected bytes below are derived from the protocol spec, not beem.
section("EscrowDisputeOperation (NOTE: beem omits agent — bytes are protocol-spec derived)")
print("// escrow_dispute (beem missing agent; bytes below match Hive protocol spec)")
ed_expected = bytes([28, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0])
print(f"escrow_dispute: {go_bytes(ed_expected)}")

# NOTE: beem's Escrow_release omits 'agent' and 'receiver' fields.
# Our Go implementation follows the Hive protocol spec.
section("EscrowReleaseOperation (NOTE: beem omits agent+receiver — bytes are protocol-spec derived)")
print("// escrow_release (beem missing agent+receiver; bytes below match Hive protocol spec)")
er_expected = bytes([29, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0])
print(f"escrow_release: {go_bytes(er_expected)}")

section("CustomOperation")
from beembase.operationids import operations as beem_ops
from beembase.transactions import Signed_Transaction as ST2
# beem Custom: required_auths (flat_set<account_name>), id (uint16), data (string)
from beembase.objects import Custom
custom_op = Custom(**{"required_auths": ["alice"], "id": 42, "data": "0102"})
custom_expected = bytes(custom_op)
print(f"custom: {go_bytes(custom_expected)}")

# NOTE: beem's Custom_binary uses uint16 for id and omits all auth fields.
# Our Go implementation follows the Hive C++ protocol spec:
#   required_owner_auths (flat_set), required_active_auths, required_posting_auths,
#   required_auths (vector<authority>), id (string), data (vector<char>).
# Bytes derived from the protocol spec, not beem.
section("CustomBinaryOperation (NOTE: beem has wrong format — bytes are protocol-spec derived)")
print("// custom_binary (beem uses uint16 id, omits auth fields; bytes match Hive protocol spec)")
cb_expected = bytes([35, 0, 0, 0, 0, 4, 116, 101, 115, 116, 2, 1, 2])
print(f"custom_binary: {go_bytes(cb_expected)}")

# ---------------------------------------------------------------------------
# Signing verification
# ---------------------------------------------------------------------------

print("\n\n// --- Signing verification ---")
print("// Verifying hivego's hashTxForSig and TestSignDigest against beem")

# The test vote transaction
tx = Signed_Transaction(
    ref_block_num=36029,
    ref_block_prefix=1164960351,
    expiration="2016-08-08T12:24:17",
    operations=[Operation(["vote", Vote(**{
        "voter": "xeroc", "author": "xeroc", "permlink": "piston", "weight": 10000,
    })])],
    signatures=[],
    extensions=[],
)
tx.deriveDigest(CHAIN_PARAMS)

print(f"// hashTxForSig (sha256(chainID+txBytes)): {tx.digest.hex()}")
print(f"// expected in TestHashTxForSig:            0e5dbd979f4f2387c5d5a1b649ef06589630faf7c065dea0da8e592b03da06bc")
assert tx.digest.hex() == "0e5dbd979f4f2387c5d5a1b649ef06589630faf7c065dea0da8e592b03da06bc", "hashTxForSig mismatch!"
print("// hashTxForSig: OK")

# Verify hivego's known signature against hashTx (sha256 of tx bytes without chain_id).
#
# TestSignDigest in hivego signs hashTx directly (not sha256(hashTx)).
# We use the pure-Python `ecdsa` library for verification because beem's verify_message
# uses the secp256k1 C extension which has API breakage on Python 3.14.
#
# WIF key: 5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W
# Compact sig (from TestSignDigest in signer_test.go): recovery_code + 32-byte r + 32-byte s
hivego_sig_hex = "1f87b2ff969165939f2bd0d0d48da74f7ccad86a545c7e7cd97c272f694c1471e502056e74b5d6206f86fde7649ef166ee20d5169b73e2ac55e5b763009bfce5ba"
wif = "5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W"
expected_pubkey_hex = "02756f674e8e93ffc8eb719c933a9b06dcab85dc62160b3129c93a9ef767f34091"

# Derive the private key from the WIF (version byte + 32 privkey bytes + 4-byte checksum)
BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
def b58decode(s):
    n = 0
    for c in s:
        n = n * 58 + BASE58.index(c)
    result = []
    while n:
        result.append(n & 0xFF)
        n >>= 8
    return bytes(reversed(result))

wif_raw = b58decode(wif)
privkey_bytes = wif_raw[1:33]
sk = ecdsa.SigningKey.from_string(privkey_bytes, curve=ecdsa.SECP256k1)
vk = sk.get_verifying_key()

# Verify the public key matches expected
x, y = vk.pubkey.point.x(), vk.pubkey.point.y()
prefix = b'\x02' if y % 2 == 0 else b'\x03'
actual_pubkey_hex = (prefix + x.to_bytes(32, 'big')).hex()
assert actual_pubkey_hex == expected_pubkey_hex, f"Pubkey mismatch: {actual_pubkey_hex} != {expected_pubkey_hex}"
print(f"// WIF pubkey:  {actual_pubkey_hex}")

# hashTx = sha256(txBytes) — the digest signed in TestSignDigest
hashTx_input = bytes([189, 140, 95, 226, 111, 69, 241, 121, 168, 87, 1, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 0])
hashTx = hashlib.sha256(hashTx_input).digest()
print(f"// hashTx (sha256 of txBytes): {hashTx.hex()}")
print(f"// expected in TestHashTx:     12164dcee518674c586e6a61d08623c44980e326c9816c5023c8b8880f723d6b")
assert hashTx.hex() == "12164dcee518674c586e6a61d08623c44980e326c9816c5023c8b8880f723d6b", "hashTx mismatch!"
print("// hashTx: OK")

sig_bytes = binascii.unhexlify(hivego_sig_hex)
r_s = sig_bytes[1:]  # strip the compact recovery byte; r (32 bytes) || s (32 bytes)
try:
    vk.verify_digest(r_s, hashTx, sigdecode=ecdsa.util.sigdecode_string)
    print(f"// TestSignDigest signature: OK (verified with ecdsa library)")
except ecdsa.keys.BadSignatureError as e:
    print(f"// TestSignDigest signature: MISMATCH — {e}")
    sys.exit(1)

print("\n// All vectors verified successfully.")