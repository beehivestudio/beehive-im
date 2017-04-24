namespace go seqsvr

service SeqSvrThrift {
    i64 AllocSid(),
    i64 GetSessionSeq(1:i64 uid),
}
