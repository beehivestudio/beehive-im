namespace go seqsvr

service SeqSvrThrift {
    i64 AllocSid(),
    i64 QuerySeqBySid(1:i64 sid),
}
