namespace go seqsvr

service SeqSvrThrift {
    i64 AllocSid(),
    i64 AllocSeq(1:i64 uid),
}
