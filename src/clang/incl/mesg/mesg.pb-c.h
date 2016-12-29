/* Generated by the protocol buffer compiler.  DO NOT EDIT! */
/* Generated from: mesg.proto */

#ifndef PROTOBUF_C_mesg_2eproto__INCLUDED
#define PROTOBUF_C_mesg_2eproto__INCLUDED

#include <protobuf-c/protobuf-c.h>

PROTOBUF_C__BEGIN_DECLS

#if PROTOBUF_C_VERSION_NUMBER < 1000000
# error This file was generated by a newer version of protoc-c which is incompatible with your libprotobuf-c headers. Please update your headers.
#elif 1000002 < PROTOBUF_C_MIN_COMPILER_VERSION
# error This file was generated by an older version of protoc-c which is incompatible with your libprotobuf-c headers. Please regenerate this file with a newer version of protoc-c.
#endif


typedef struct _MesgOnlineReq MesgOnlineReq;
typedef struct _MesgOnlineAck MesgOnlineAck;
typedef struct _MesgJoinReq MesgJoinReq;
typedef struct _MesgJoinAck MesgJoinAck;
typedef struct _MesgUnjoinReq MesgUnjoinReq;
typedef struct _MesgRoom MesgRoom;
typedef struct _MesgRoomAck MesgRoomAck;
typedef struct _MesgGroup MesgGroup;
typedef struct _MesgGroupAck MesgGroupAck;
typedef struct _MesgLsnRpt MesgLsnRpt;
typedef struct _MesgKickReq MesgKickReq;


/* --- enums --- */


/* --- messages --- */

struct  _MesgOnlineReq
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t sid;
  char *token;
  char *app;
  char *version;
  protobuf_c_boolean has_terminal;
  uint32_t terminal;
};
#define MESG_ONLINE_REQ__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_online_req__descriptor) \
    , 0, 0, NULL, NULL, NULL, 0,0 }


struct  _MesgOnlineAck
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t sid;
  char *app;
  char *version;
  protobuf_c_boolean has_terminal;
  uint32_t terminal;
  uint32_t code;
  char *errmsg;
};
#define MESG_ONLINE_ACK__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_online_ack__descriptor) \
    , 0, 0, NULL, NULL, 0,0, 0, NULL }


struct  _MesgJoinReq
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t rid;
  char *token;
};
#define MESG_JOIN_REQ__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_join_req__descriptor) \
    , 0, 0, NULL }


struct  _MesgJoinAck
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t rid;
  uint32_t gid;
  uint32_t code;
  char *errmsg;
};
#define MESG_JOIN_ACK__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_join_ack__descriptor) \
    , 0, 0, 0, 0, NULL }


struct  _MesgUnjoinReq
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t rid;
};
#define MESG_UNJOIN_REQ__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_unjoin_req__descriptor) \
    , 0, 0 }


struct  _MesgRoom
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t rid;
  uint32_t gid;
  uint32_t level;
  uint64_t time;
  char *text;
  protobuf_c_boolean has_data;
  ProtobufCBinaryData data;
};
#define MESG_ROOM__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_room__descriptor) \
    , 0, 0, 0, 0, 0, NULL, 0,{0,NULL} }


struct  _MesgRoomAck
{
  ProtobufCMessage base;
  uint32_t code;
  char *errmsg;
};
#define MESG_ROOM_ACK__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_room_ack__descriptor) \
    , 0, NULL }


struct  _MesgGroup
{
  ProtobufCMessage base;
  uint64_t uid;
  uint64_t gid;
  uint32_t level;
  uint64_t time;
  char *text;
  protobuf_c_boolean has_data;
  ProtobufCBinaryData data;
};
#define MESG_GROUP__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_group__descriptor) \
    , 0, 0, 0, 0, NULL, 0,{0,NULL} }


struct  _MesgGroupAck
{
  ProtobufCMessage base;
  uint32_t code;
  char *errmsg;
};
#define MESG_GROUP_ACK__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_group_ack__descriptor) \
    , 0, NULL }


struct  _MesgLsnRpt
{
  ProtobufCMessage base;
  uint32_t nid;
  char *nation;
  char *name;
  char *ipaddr;
  uint32_t port;
};
#define MESG_LSN_RPT__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_lsn_rpt__descriptor) \
    , 0, NULL, NULL, NULL, 0 }


struct  _MesgKickReq
{
  ProtobufCMessage base;
  uint32_t code;
  char *errmsg;
};
#define MESG_KICK_REQ__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&mesg_kick_req__descriptor) \
    , 0, NULL }


/* MesgOnlineReq methods */
void   mesg_online_req__init
                     (MesgOnlineReq         *message);
size_t mesg_online_req__get_packed_size
                     (const MesgOnlineReq   *message);
size_t mesg_online_req__pack
                     (const MesgOnlineReq   *message,
                      uint8_t             *out);
size_t mesg_online_req__pack_to_buffer
                     (const MesgOnlineReq   *message,
                      ProtobufCBuffer     *buffer);
MesgOnlineReq *
       mesg_online_req__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_online_req__free_unpacked
                     (MesgOnlineReq *message,
                      ProtobufCAllocator *allocator);
/* MesgOnlineAck methods */
void   mesg_online_ack__init
                     (MesgOnlineAck         *message);
size_t mesg_online_ack__get_packed_size
                     (const MesgOnlineAck   *message);
size_t mesg_online_ack__pack
                     (const MesgOnlineAck   *message,
                      uint8_t             *out);
size_t mesg_online_ack__pack_to_buffer
                     (const MesgOnlineAck   *message,
                      ProtobufCBuffer     *buffer);
MesgOnlineAck *
       mesg_online_ack__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_online_ack__free_unpacked
                     (MesgOnlineAck *message,
                      ProtobufCAllocator *allocator);
/* MesgJoinReq methods */
void   mesg_join_req__init
                     (MesgJoinReq         *message);
size_t mesg_join_req__get_packed_size
                     (const MesgJoinReq   *message);
size_t mesg_join_req__pack
                     (const MesgJoinReq   *message,
                      uint8_t             *out);
size_t mesg_join_req__pack_to_buffer
                     (const MesgJoinReq   *message,
                      ProtobufCBuffer     *buffer);
MesgJoinReq *
       mesg_join_req__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_join_req__free_unpacked
                     (MesgJoinReq *message,
                      ProtobufCAllocator *allocator);
/* MesgJoinAck methods */
void   mesg_join_ack__init
                     (MesgJoinAck         *message);
size_t mesg_join_ack__get_packed_size
                     (const MesgJoinAck   *message);
size_t mesg_join_ack__pack
                     (const MesgJoinAck   *message,
                      uint8_t             *out);
size_t mesg_join_ack__pack_to_buffer
                     (const MesgJoinAck   *message,
                      ProtobufCBuffer     *buffer);
MesgJoinAck *
       mesg_join_ack__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_join_ack__free_unpacked
                     (MesgJoinAck *message,
                      ProtobufCAllocator *allocator);
/* MesgUnjoinReq methods */
void   mesg_unjoin_req__init
                     (MesgUnjoinReq         *message);
size_t mesg_unjoin_req__get_packed_size
                     (const MesgUnjoinReq   *message);
size_t mesg_unjoin_req__pack
                     (const MesgUnjoinReq   *message,
                      uint8_t             *out);
size_t mesg_unjoin_req__pack_to_buffer
                     (const MesgUnjoinReq   *message,
                      ProtobufCBuffer     *buffer);
MesgUnjoinReq *
       mesg_unjoin_req__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_unjoin_req__free_unpacked
                     (MesgUnjoinReq *message,
                      ProtobufCAllocator *allocator);
/* MesgRoom methods */
void   mesg_room__init
                     (MesgRoom         *message);
size_t mesg_room__get_packed_size
                     (const MesgRoom   *message);
size_t mesg_room__pack
                     (const MesgRoom   *message,
                      uint8_t             *out);
size_t mesg_room__pack_to_buffer
                     (const MesgRoom   *message,
                      ProtobufCBuffer     *buffer);
MesgRoom *
       mesg_room__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_room__free_unpacked
                     (MesgRoom *message,
                      ProtobufCAllocator *allocator);
/* MesgRoomAck methods */
void   mesg_room_ack__init
                     (MesgRoomAck         *message);
size_t mesg_room_ack__get_packed_size
                     (const MesgRoomAck   *message);
size_t mesg_room_ack__pack
                     (const MesgRoomAck   *message,
                      uint8_t             *out);
size_t mesg_room_ack__pack_to_buffer
                     (const MesgRoomAck   *message,
                      ProtobufCBuffer     *buffer);
MesgRoomAck *
       mesg_room_ack__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_room_ack__free_unpacked
                     (MesgRoomAck *message,
                      ProtobufCAllocator *allocator);
/* MesgGroup methods */
void   mesg_group__init
                     (MesgGroup         *message);
size_t mesg_group__get_packed_size
                     (const MesgGroup   *message);
size_t mesg_group__pack
                     (const MesgGroup   *message,
                      uint8_t             *out);
size_t mesg_group__pack_to_buffer
                     (const MesgGroup   *message,
                      ProtobufCBuffer     *buffer);
MesgGroup *
       mesg_group__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_group__free_unpacked
                     (MesgGroup *message,
                      ProtobufCAllocator *allocator);
/* MesgGroupAck methods */
void   mesg_group_ack__init
                     (MesgGroupAck         *message);
size_t mesg_group_ack__get_packed_size
                     (const MesgGroupAck   *message);
size_t mesg_group_ack__pack
                     (const MesgGroupAck   *message,
                      uint8_t             *out);
size_t mesg_group_ack__pack_to_buffer
                     (const MesgGroupAck   *message,
                      ProtobufCBuffer     *buffer);
MesgGroupAck *
       mesg_group_ack__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_group_ack__free_unpacked
                     (MesgGroupAck *message,
                      ProtobufCAllocator *allocator);
/* MesgLsnRpt methods */
void   mesg_lsn_rpt__init
                     (MesgLsnRpt         *message);
size_t mesg_lsn_rpt__get_packed_size
                     (const MesgLsnRpt   *message);
size_t mesg_lsn_rpt__pack
                     (const MesgLsnRpt   *message,
                      uint8_t             *out);
size_t mesg_lsn_rpt__pack_to_buffer
                     (const MesgLsnRpt   *message,
                      ProtobufCBuffer     *buffer);
MesgLsnRpt *
       mesg_lsn_rpt__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_lsn_rpt__free_unpacked
                     (MesgLsnRpt *message,
                      ProtobufCAllocator *allocator);
/* MesgKickReq methods */
void   mesg_kick_req__init
                     (MesgKickReq         *message);
size_t mesg_kick_req__get_packed_size
                     (const MesgKickReq   *message);
size_t mesg_kick_req__pack
                     (const MesgKickReq   *message,
                      uint8_t             *out);
size_t mesg_kick_req__pack_to_buffer
                     (const MesgKickReq   *message,
                      ProtobufCBuffer     *buffer);
MesgKickReq *
       mesg_kick_req__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   mesg_kick_req__free_unpacked
                     (MesgKickReq *message,
                      ProtobufCAllocator *allocator);
/* --- per-message closures --- */

typedef void (*MesgOnlineReq_Closure)
                 (const MesgOnlineReq *message,
                  void *closure_data);
typedef void (*MesgOnlineAck_Closure)
                 (const MesgOnlineAck *message,
                  void *closure_data);
typedef void (*MesgJoinReq_Closure)
                 (const MesgJoinReq *message,
                  void *closure_data);
typedef void (*MesgJoinAck_Closure)
                 (const MesgJoinAck *message,
                  void *closure_data);
typedef void (*MesgUnjoinReq_Closure)
                 (const MesgUnjoinReq *message,
                  void *closure_data);
typedef void (*MesgRoom_Closure)
                 (const MesgRoom *message,
                  void *closure_data);
typedef void (*MesgRoomAck_Closure)
                 (const MesgRoomAck *message,
                  void *closure_data);
typedef void (*MesgGroup_Closure)
                 (const MesgGroup *message,
                  void *closure_data);
typedef void (*MesgGroupAck_Closure)
                 (const MesgGroupAck *message,
                  void *closure_data);
typedef void (*MesgLsnRpt_Closure)
                 (const MesgLsnRpt *message,
                  void *closure_data);
typedef void (*MesgKickReq_Closure)
                 (const MesgKickReq *message,
                  void *closure_data);

/* --- services --- */


/* --- descriptors --- */

extern const ProtobufCMessageDescriptor mesg_online_req__descriptor;
extern const ProtobufCMessageDescriptor mesg_online_ack__descriptor;
extern const ProtobufCMessageDescriptor mesg_join_req__descriptor;
extern const ProtobufCMessageDescriptor mesg_join_ack__descriptor;
extern const ProtobufCMessageDescriptor mesg_unjoin_req__descriptor;
extern const ProtobufCMessageDescriptor mesg_room__descriptor;
extern const ProtobufCMessageDescriptor mesg_room_ack__descriptor;
extern const ProtobufCMessageDescriptor mesg_group__descriptor;
extern const ProtobufCMessageDescriptor mesg_group_ack__descriptor;
extern const ProtobufCMessageDescriptor mesg_lsn_rpt__descriptor;
extern const ProtobufCMessageDescriptor mesg_kick_req__descriptor;

PROTOBUF_C__END_DECLS


#endif  /* PROTOBUF_C_mesg_2eproto__INCLUDED */
