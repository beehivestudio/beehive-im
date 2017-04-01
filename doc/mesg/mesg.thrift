namespace go seqsvr

struct MesgAllocSid { }

service SeqSvrThrift {
    i64 AllocSid(),
}
