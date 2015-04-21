// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/fastsha256"
	"github.com/btcsuite/golangcrypto/ripemd160"
)

// An opcode defines the information related to a txscript opcode.
// opfunc if present is the function to call to perform the opcode on
// the script. The current script is passed in as a slice with the firs
// member being the opcode itself.
type opcode struct {
	value  byte
	name   string
	length int
	opfunc func(*parsedOpcode, *Engine) error
}

// These constants are the values of the official opcodes used on the btc wiki,
// in bitcoin core and in most if not all other references and software related
// to handling BTC scripts.
const (
	OP_0                   = 0x00 // 0
	OP_FALSE               = 0x00 // 0 - AKA OP_0
	OP_DATA_1              = 0x01 // 1
	OP_DATA_2              = 0x02 // 2
	OP_DATA_3              = 0x03 // 3
	OP_DATA_4              = 0x04 // 4
	OP_DATA_5              = 0x05 // 5
	OP_DATA_6              = 0x06 // 6
	OP_DATA_7              = 0x07 // 7
	OP_DATA_8              = 0x08 // 8
	OP_DATA_9              = 0x09 // 9
	OP_DATA_10             = 0x0a // 10
	OP_DATA_11             = 0x0b // 11
	OP_DATA_12             = 0x0c // 12
	OP_DATA_13             = 0x0d // 13
	OP_DATA_14             = 0x0e // 14
	OP_DATA_15             = 0x0f // 15
	OP_DATA_16             = 0x10 // 16
	OP_DATA_17             = 0x11 // 17
	OP_DATA_18             = 0x12 // 18
	OP_DATA_19             = 0x13 // 19
	OP_DATA_20             = 0x14 // 20
	OP_DATA_21             = 0x15 // 21
	OP_DATA_22             = 0x16 // 22
	OP_DATA_23             = 0x17 // 23
	OP_DATA_24             = 0x18 // 24
	OP_DATA_25             = 0x19 // 25
	OP_DATA_26             = 0x1a // 26
	OP_DATA_27             = 0x1b // 27
	OP_DATA_28             = 0x1c // 28
	OP_DATA_29             = 0x1d // 29
	OP_DATA_30             = 0x1e // 30
	OP_DATA_31             = 0x1f // 31
	OP_DATA_32             = 0x20 // 32
	OP_DATA_33             = 0x21 // 33
	OP_DATA_34             = 0x22 // 34
	OP_DATA_35             = 0x23 // 35
	OP_DATA_36             = 0x24 // 36
	OP_DATA_37             = 0x25 // 37
	OP_DATA_38             = 0x26 // 38
	OP_DATA_39             = 0x27 // 39
	OP_DATA_40             = 0x28 // 40
	OP_DATA_41             = 0x29 // 41
	OP_DATA_42             = 0x2a // 42
	OP_DATA_43             = 0x2b // 43
	OP_DATA_44             = 0x2c // 44
	OP_DATA_45             = 0x2d // 45
	OP_DATA_46             = 0x2e // 46
	OP_DATA_47             = 0x2f // 47
	OP_DATA_48             = 0x30 // 48
	OP_DATA_49             = 0x31 // 49
	OP_DATA_50             = 0x32 // 50
	OP_DATA_51             = 0x33 // 51
	OP_DATA_52             = 0x34 // 52
	OP_DATA_53             = 0x35 // 53
	OP_DATA_54             = 0x36 // 54
	OP_DATA_55             = 0x37 // 55
	OP_DATA_56             = 0x38 // 56
	OP_DATA_57             = 0x39 // 57
	OP_DATA_58             = 0x3a // 58
	OP_DATA_59             = 0x3b // 59
	OP_DATA_60             = 0x3c // 60
	OP_DATA_61             = 0x3d // 61
	OP_DATA_62             = 0x3e // 62
	OP_DATA_63             = 0x3f // 63
	OP_DATA_64             = 0x40 // 64
	OP_DATA_65             = 0x41 // 65
	OP_DATA_66             = 0x42 // 66
	OP_DATA_67             = 0x43 // 67
	OP_DATA_68             = 0x44 // 68
	OP_DATA_69             = 0x45 // 69
	OP_DATA_70             = 0x46 // 70
	OP_DATA_71             = 0x47 // 71
	OP_DATA_72             = 0x48 // 72
	OP_DATA_73             = 0x49 // 73
	OP_DATA_74             = 0x4a // 74
	OP_DATA_75             = 0x4b // 75
	OP_PUSHDATA1           = 0x4c // 76
	OP_PUSHDATA2           = 0x4d // 77
	OP_PUSHDATA4           = 0x4e // 78
	OP_1NEGATE             = 0x4f // 79
	OP_RESERVED            = 0x50 // 80
	OP_1                   = 0x51 // 81 - AKA OP_TRUE
	OP_TRUE                = 0x51 // 81
	OP_2                   = 0x52 // 82
	OP_3                   = 0x53 // 83
	OP_4                   = 0x54 // 84
	OP_5                   = 0x55 // 85
	OP_6                   = 0x56 // 86
	OP_7                   = 0x57 // 87
	OP_8                   = 0x58 // 88
	OP_9                   = 0x59 // 89
	OP_10                  = 0x5a // 90
	OP_11                  = 0x5b // 91
	OP_12                  = 0x5c // 92
	OP_13                  = 0x5d // 93
	OP_14                  = 0x5e // 94
	OP_15                  = 0x5f // 95
	OP_16                  = 0x60 // 96
	OP_NOP                 = 0x61 // 97
	OP_VER                 = 0x62 // 98
	OP_IF                  = 0x63 // 99
	OP_NOTIF               = 0x64 // 100
	OP_VERIF               = 0x65 // 101
	OP_VERNOTIF            = 0x66 // 102
	OP_ELSE                = 0x67 // 103
	OP_ENDIF               = 0x68 // 104
	OP_VERIFY              = 0x69 // 105
	OP_RETURN              = 0x6a // 106
	OP_TOALTSTACK          = 0x6b // 107
	OP_FROMALTSTACK        = 0x6c // 108
	OP_2DROP               = 0x6d // 109
	OP_2DUP                = 0x6e // 110
	OP_3DUP                = 0x6f // 111
	OP_2OVER               = 0x70 // 112
	OP_2ROT                = 0x71 // 113
	OP_2SWAP               = 0x72 // 114
	OP_IFDUP               = 0x73 // 115
	OP_DEPTH               = 0x74 // 116
	OP_DROP                = 0x75 // 117
	OP_DUP                 = 0x76 // 118
	OP_NIP                 = 0x77 // 119
	OP_OVER                = 0x78 // 120
	OP_PICK                = 0x79 // 121
	OP_ROLL                = 0x7a // 122
	OP_ROT                 = 0x7b // 123
	OP_SWAP                = 0x7c // 124
	OP_TUCK                = 0x7d // 125
	OP_CAT                 = 0x7e // 126
	OP_SUBSTR              = 0x7f // 127
	OP_LEFT                = 0x80 // 128
	OP_RIGHT               = 0x81 // 129
	OP_SIZE                = 0x82 // 130
	OP_INVERT              = 0x83 // 131
	OP_AND                 = 0x84 // 132
	OP_OR                  = 0x85 // 133
	OP_XOR                 = 0x86 // 134
	OP_EQUAL               = 0x87 // 135
	OP_EQUALVERIFY         = 0x88 // 136
	OP_RESERVED1           = 0x89 // 137
	OP_RESERVED2           = 0x8a // 138
	OP_1ADD                = 0x8b // 139
	OP_1SUB                = 0x8c // 140
	OP_2MUL                = 0x8d // 141
	OP_2DIV                = 0x8e // 142
	OP_NEGATE              = 0x8f // 143
	OP_ABS                 = 0x90 // 144
	OP_NOT                 = 0x91 // 145
	OP_0NOTEQUAL           = 0x92 // 146
	OP_ADD                 = 0x93 // 147
	OP_SUB                 = 0x94 // 148
	OP_MUL                 = 0x95 // 149
	OP_DIV                 = 0x96 // 150
	OP_MOD                 = 0x97 // 151
	OP_LSHIFT              = 0x98 // 152
	OP_RSHIFT              = 0x99 // 153
	OP_BOOLAND             = 0x9a // 154
	OP_BOOLOR              = 0x9b // 155
	OP_NUMEQUAL            = 0x9c // 156
	OP_NUMEQUALVERIFY      = 0x9d // 157
	OP_NUMNOTEQUAL         = 0x9e // 158
	OP_LESSTHAN            = 0x9f // 159
	OP_GREATERTHAN         = 0xa0 // 160
	OP_LESSTHANOREQUAL     = 0xa1 // 161
	OP_GREATERTHANOREQUAL  = 0xa2 // 162
	OP_MIN                 = 0xa3 // 163
	OP_MAX                 = 0xa4 // 164
	OP_WITHIN              = 0xa5 // 165
	OP_RIPEMD160           = 0xa6 // 166
	OP_SHA1                = 0xa7 // 167
	OP_SHA256              = 0xa8 // 168
	OP_HASH160             = 0xa9 // 169
	OP_HASH256             = 0xaa // 170
	OP_CODESEPARATOR       = 0xab // 171
	OP_CHECKSIG            = 0xac // 172
	OP_CHECKSIGVERIFY      = 0xad // 173
	OP_CHECKMULTISIG       = 0xae // 174
	OP_CHECKMULTISIGVERIFY = 0xaf // 175
	OP_NOP1                = 0xb0 // 176
	OP_NOP2                = 0xb1 // 177
	OP_NOP3                = 0xb2 // 178
	OP_NOP4                = 0xb3 // 179
	OP_NOP5                = 0xb4 // 180
	OP_NOP6                = 0xb5 // 181
	OP_NOP7                = 0xb6 // 182
	OP_NOP8                = 0xb7 // 183
	OP_NOP9                = 0xb8 // 184
	OP_NOP10               = 0xb9 // 185
	OP_UNKNOWN186          = 0xba // 186
	OP_UNKNOWN187          = 0xbb // 187
	OP_UNKNOWN188          = 0xbc // 188
	OP_UNKNOWN189          = 0xbd // 189
	OP_UNKNOWN190          = 0xbe // 190
	OP_UNKNOWN191          = 0xbf // 191
	OP_UNKNOWN192          = 0xc0 // 192
	OP_UNKNOWN193          = 0xc1 // 193
	OP_UNKNOWN194          = 0xc2 // 194
	OP_UNKNOWN195          = 0xc3 // 195
	OP_UNKNOWN196          = 0xc4 // 196
	OP_UNKNOWN197          = 0xc5 // 197
	OP_UNKNOWN198          = 0xc6 // 198
	OP_UNKNOWN199          = 0xc7 // 199
	OP_UNKNOWN200          = 0xc8 // 200
	OP_UNKNOWN201          = 0xc9 // 201
	OP_UNKNOWN202          = 0xca // 202
	OP_UNKNOWN203          = 0xcb // 203
	OP_UNKNOWN204          = 0xcc // 204
	OP_UNKNOWN205          = 0xcd // 205
	OP_UNKNOWN206          = 0xce // 206
	OP_UNKNOWN207          = 0xcf // 207
	OP_UNKNOWN208          = 0xd0 // 208
	OP_UNKNOWN209          = 0xd1 // 209
	OP_UNKNOWN210          = 0xd2 // 210
	OP_UNKNOWN211          = 0xd3 // 211
	OP_UNKNOWN212          = 0xd4 // 212
	OP_UNKNOWN213          = 0xd5 // 213
	OP_UNKNOWN214          = 0xd6 // 214
	OP_UNKNOWN215          = 0xd7 // 215
	OP_UNKNOWN216          = 0xd8 // 216
	OP_UNKNOWN217          = 0xd9 // 217
	OP_UNKNOWN218          = 0xda // 218
	OP_UNKNOWN219          = 0xdb // 219
	OP_UNKNOWN220          = 0xdc // 220
	OP_UNKNOWN221          = 0xdd // 221
	OP_UNKNOWN222          = 0xde // 222
	OP_UNKNOWN223          = 0xdf // 223
	OP_UNKNOWN224          = 0xe0 // 224
	OP_UNKNOWN225          = 0xe1 // 225
	OP_UNKNOWN226          = 0xe2 // 226
	OP_UNKNOWN227          = 0xe3 // 227
	OP_UNKNOWN228          = 0xe4 // 228
	OP_UNKNOWN229          = 0xe5 // 229
	OP_UNKNOWN230          = 0xe6 // 230
	OP_UNKNOWN231          = 0xe7 // 231
	OP_UNKNOWN232          = 0xe8 // 232
	OP_UNKNOWN233          = 0xe9 // 233
	OP_UNKNOWN234          = 0xea // 234
	OP_UNKNOWN235          = 0xeb // 235
	OP_UNKNOWN236          = 0xec // 236
	OP_UNKNOWN237          = 0xed // 237
	OP_UNKNOWN238          = 0xee // 238
	OP_UNKNOWN239          = 0xef // 239
	OP_UNKNOWN240          = 0xf0 // 240
	OP_UNKNOWN241          = 0xf1 // 241
	OP_UNKNOWN242          = 0xf2 // 242
	OP_UNKNOWN243          = 0xf3 // 243
	OP_UNKNOWN244          = 0xf4 // 244
	OP_UNKNOWN245          = 0xf5 // 245
	OP_UNKNOWN246          = 0xf6 // 246
	OP_UNKNOWN247          = 0xf7 // 247
	OP_UNKNOWN248          = 0xf8 // 248
	OP_SMALLDATA           = 0xf9 // 249 - bitcoin core internal
	OP_SMALLINTEGER        = 0xfa // 250 - bitcoin core internal
	OP_PUBKEYS             = 0xfb // 251 - bitcoin core internal
	OP_UNKNOWN252          = 0xfc // 252
	OP_PUBKEYHASH          = 0xfd // 253 - bitcoin core internal
	OP_PUBKEY              = 0xfe // 254 - bitcoin core internal
	OP_INVALIDOPCODE       = 0xff // 255 - bitcoin core internal
)

// conditional execution constants
const (
	OpCondFalse = 0
	OpCondTrue  = 1
	OpCondSkip  = 2
)

// opcodeArray holds details about all possible opcodes such as how many bytes
// the opcode and any associated data should take, its human-readable name, and
// the handler function.
var opcodeArray = [256]opcode{
	// Data push opcodes.
	OP_FALSE:     {OP_FALSE, "OP_0", 1, opcodeFalse},
	OP_DATA_1:    {OP_DATA_1, "OP_DATA_1", 2, opcodePushData},
	OP_DATA_2:    {OP_DATA_2, "OP_DATA_2", 3, opcodePushData},
	OP_DATA_3:    {OP_DATA_3, "OP_DATA_3", 4, opcodePushData},
	OP_DATA_4:    {OP_DATA_4, "OP_DATA_4", 5, opcodePushData},
	OP_DATA_5:    {OP_DATA_5, "OP_DATA_5", 6, opcodePushData},
	OP_DATA_6:    {OP_DATA_6, "OP_DATA_6", 7, opcodePushData},
	OP_DATA_7:    {OP_DATA_7, "OP_DATA_7", 8, opcodePushData},
	OP_DATA_8:    {OP_DATA_8, "OP_DATA_8", 9, opcodePushData},
	OP_DATA_9:    {OP_DATA_9, "OP_DATA_9", 10, opcodePushData},
	OP_DATA_10:   {OP_DATA_10, "OP_DATA_10", 11, opcodePushData},
	OP_DATA_11:   {OP_DATA_11, "OP_DATA_11", 12, opcodePushData},
	OP_DATA_12:   {OP_DATA_12, "OP_DATA_12", 13, opcodePushData},
	OP_DATA_13:   {OP_DATA_13, "OP_DATA_13", 14, opcodePushData},
	OP_DATA_14:   {OP_DATA_14, "OP_DATA_14", 15, opcodePushData},
	OP_DATA_15:   {OP_DATA_15, "OP_DATA_15", 16, opcodePushData},
	OP_DATA_16:   {OP_DATA_16, "OP_DATA_16", 17, opcodePushData},
	OP_DATA_17:   {OP_DATA_17, "OP_DATA_17", 18, opcodePushData},
	OP_DATA_18:   {OP_DATA_18, "OP_DATA_18", 19, opcodePushData},
	OP_DATA_19:   {OP_DATA_19, "OP_DATA_19", 20, opcodePushData},
	OP_DATA_20:   {OP_DATA_20, "OP_DATA_20", 21, opcodePushData},
	OP_DATA_21:   {OP_DATA_21, "OP_DATA_21", 22, opcodePushData},
	OP_DATA_22:   {OP_DATA_22, "OP_DATA_22", 23, opcodePushData},
	OP_DATA_23:   {OP_DATA_23, "OP_DATA_23", 24, opcodePushData},
	OP_DATA_24:   {OP_DATA_24, "OP_DATA_24", 25, opcodePushData},
	OP_DATA_25:   {OP_DATA_25, "OP_DATA_25", 26, opcodePushData},
	OP_DATA_26:   {OP_DATA_26, "OP_DATA_26", 27, opcodePushData},
	OP_DATA_27:   {OP_DATA_27, "OP_DATA_27", 28, opcodePushData},
	OP_DATA_28:   {OP_DATA_28, "OP_DATA_28", 29, opcodePushData},
	OP_DATA_29:   {OP_DATA_29, "OP_DATA_29", 30, opcodePushData},
	OP_DATA_30:   {OP_DATA_30, "OP_DATA_30", 31, opcodePushData},
	OP_DATA_31:   {OP_DATA_31, "OP_DATA_31", 32, opcodePushData},
	OP_DATA_32:   {OP_DATA_32, "OP_DATA_32", 33, opcodePushData},
	OP_DATA_33:   {OP_DATA_33, "OP_DATA_33", 34, opcodePushData},
	OP_DATA_34:   {OP_DATA_34, "OP_DATA_34", 35, opcodePushData},
	OP_DATA_35:   {OP_DATA_35, "OP_DATA_35", 36, opcodePushData},
	OP_DATA_36:   {OP_DATA_36, "OP_DATA_36", 37, opcodePushData},
	OP_DATA_37:   {OP_DATA_37, "OP_DATA_37", 38, opcodePushData},
	OP_DATA_38:   {OP_DATA_38, "OP_DATA_38", 39, opcodePushData},
	OP_DATA_39:   {OP_DATA_39, "OP_DATA_39", 40, opcodePushData},
	OP_DATA_40:   {OP_DATA_40, "OP_DATA_40", 41, opcodePushData},
	OP_DATA_41:   {OP_DATA_41, "OP_DATA_41", 42, opcodePushData},
	OP_DATA_42:   {OP_DATA_42, "OP_DATA_42", 43, opcodePushData},
	OP_DATA_43:   {OP_DATA_43, "OP_DATA_43", 44, opcodePushData},
	OP_DATA_44:   {OP_DATA_44, "OP_DATA_44", 45, opcodePushData},
	OP_DATA_45:   {OP_DATA_45, "OP_DATA_45", 46, opcodePushData},
	OP_DATA_46:   {OP_DATA_46, "OP_DATA_46", 47, opcodePushData},
	OP_DATA_47:   {OP_DATA_47, "OP_DATA_47", 48, opcodePushData},
	OP_DATA_48:   {OP_DATA_48, "OP_DATA_48", 49, opcodePushData},
	OP_DATA_49:   {OP_DATA_49, "OP_DATA_49", 50, opcodePushData},
	OP_DATA_50:   {OP_DATA_50, "OP_DATA_50", 51, opcodePushData},
	OP_DATA_51:   {OP_DATA_51, "OP_DATA_51", 52, opcodePushData},
	OP_DATA_52:   {OP_DATA_52, "OP_DATA_52", 53, opcodePushData},
	OP_DATA_53:   {OP_DATA_53, "OP_DATA_53", 54, opcodePushData},
	OP_DATA_54:   {OP_DATA_54, "OP_DATA_54", 55, opcodePushData},
	OP_DATA_55:   {OP_DATA_55, "OP_DATA_55", 56, opcodePushData},
	OP_DATA_56:   {OP_DATA_56, "OP_DATA_56", 57, opcodePushData},
	OP_DATA_57:   {OP_DATA_57, "OP_DATA_57", 58, opcodePushData},
	OP_DATA_58:   {OP_DATA_58, "OP_DATA_58", 59, opcodePushData},
	OP_DATA_59:   {OP_DATA_59, "OP_DATA_59", 60, opcodePushData},
	OP_DATA_60:   {OP_DATA_60, "OP_DATA_60", 61, opcodePushData},
	OP_DATA_61:   {OP_DATA_61, "OP_DATA_61", 62, opcodePushData},
	OP_DATA_62:   {OP_DATA_62, "OP_DATA_62", 63, opcodePushData},
	OP_DATA_63:   {OP_DATA_63, "OP_DATA_63", 64, opcodePushData},
	OP_DATA_64:   {OP_DATA_64, "OP_DATA_64", 65, opcodePushData},
	OP_DATA_65:   {OP_DATA_65, "OP_DATA_65", 66, opcodePushData},
	OP_DATA_66:   {OP_DATA_66, "OP_DATA_66", 67, opcodePushData},
	OP_DATA_67:   {OP_DATA_67, "OP_DATA_67", 68, opcodePushData},
	OP_DATA_68:   {OP_DATA_68, "OP_DATA_68", 69, opcodePushData},
	OP_DATA_69:   {OP_DATA_69, "OP_DATA_69", 70, opcodePushData},
	OP_DATA_70:   {OP_DATA_70, "OP_DATA_70", 71, opcodePushData},
	OP_DATA_71:   {OP_DATA_71, "OP_DATA_71", 72, opcodePushData},
	OP_DATA_72:   {OP_DATA_72, "OP_DATA_72", 73, opcodePushData},
	OP_DATA_73:   {OP_DATA_73, "OP_DATA_73", 74, opcodePushData},
	OP_DATA_74:   {OP_DATA_74, "OP_DATA_74", 75, opcodePushData},
	OP_DATA_75:   {OP_DATA_75, "OP_DATA_75", 76, opcodePushData},
	OP_PUSHDATA1: {OP_PUSHDATA1, "OP_PUSHDATA1", -1, opcodePushData},
	OP_PUSHDATA2: {OP_PUSHDATA2, "OP_PUSHDATA2", -2, opcodePushData},
	OP_PUSHDATA4: {OP_PUSHDATA4, "OP_PUSHDATA4", -4, opcodePushData},
	OP_1NEGATE:   {OP_1NEGATE, "OP_1NEGATE", 1, opcode1Negate},
	OP_RESERVED:  {OP_RESERVED, "OP_RESERVED", 1, opcodeReserved},
	OP_TRUE:      {OP_TRUE, "OP_1", 1, opcodeN},
	OP_2:         {OP_2, "OP_2", 1, opcodeN},
	OP_3:         {OP_3, "OP_3", 1, opcodeN},
	OP_4:         {OP_4, "OP_4", 1, opcodeN},
	OP_5:         {OP_5, "OP_5", 1, opcodeN},
	OP_6:         {OP_6, "OP_6", 1, opcodeN},
	OP_7:         {OP_7, "OP_7", 1, opcodeN},
	OP_8:         {OP_8, "OP_8", 1, opcodeN},
	OP_9:         {OP_9, "OP_9", 1, opcodeN},
	OP_10:        {OP_10, "OP_10", 1, opcodeN},
	OP_11:        {OP_11, "OP_11", 1, opcodeN},
	OP_12:        {OP_12, "OP_12", 1, opcodeN},
	OP_13:        {OP_13, "OP_13", 1, opcodeN},
	OP_14:        {OP_14, "OP_14", 1, opcodeN},
	OP_15:        {OP_15, "OP_15", 1, opcodeN},
	OP_16:        {OP_16, "OP_16", 1, opcodeN},

	// Control opcodes.
	OP_NOP:      {OP_NOP, "OP_NOP", 1, opcodeNop},
	OP_VER:      {OP_VER, "OP_VER", 1, opcodeReserved},
	OP_IF:       {OP_IF, "OP_IF", 1, opcodeIf},
	OP_NOTIF:    {OP_NOTIF, "OP_NOTIF", 1, opcodeNotIf},
	OP_VERIF:    {OP_VERIF, "OP_VERIF", 1, opcodeReserved},
	OP_VERNOTIF: {OP_VERNOTIF, "OP_VERNOTIF", 1, opcodeReserved},
	OP_ELSE:     {OP_ELSE, "OP_ELSE", 1, opcodeElse},
	OP_ENDIF:    {OP_ENDIF, "OP_ENDIF", 1, opcodeEndif},
	OP_VERIFY:   {OP_VERIFY, "OP_VERIFY", 1, opcodeVerify},
	OP_RETURN:   {OP_RETURN, "OP_RETURN", 1, opcodeReturn},

	// Stack opcodes.
	OP_TOALTSTACK:   {OP_TOALTSTACK, "OP_TOALTSTACK", 1, opcodeToAltStack},
	OP_FROMALTSTACK: {OP_FROMALTSTACK, "OP_FROMALTSTACK", 1, opcodeFromAltStack},
	OP_2DROP:        {OP_2DROP, "OP_2DROP", 1, opcode2Drop},
	OP_2DUP:         {OP_2DUP, "OP_2DUP", 1, opcode2Dup},
	OP_3DUP:         {OP_3DUP, "OP_3DUP", 1, opcode3Dup},
	OP_2OVER:        {OP_2OVER, "OP_2OVER", 1, opcode2Over},
	OP_2ROT:         {OP_2ROT, "OP_2ROT", 1, opcode2Rot},
	OP_2SWAP:        {OP_2SWAP, "OP_2SWAP", 1, opcode2Swap},
	OP_IFDUP:        {OP_IFDUP, "OP_IFDUP", 1, opcodeIfDup},
	OP_DEPTH:        {OP_DEPTH, "OP_DEPTH", 1, opcodeDepth},
	OP_DROP:         {OP_DROP, "OP_DROP", 1, opcodeDrop},
	OP_DUP:          {OP_DUP, "OP_DUP", 1, opcodeDup},
	OP_NIP:          {OP_NIP, "OP_NIP", 1, opcodeNip},
	OP_OVER:         {OP_OVER, "OP_OVER", 1, opcodeOver},
	OP_PICK:         {OP_PICK, "OP_PICK", 1, opcodePick},
	OP_ROLL:         {OP_ROLL, "OP_ROLL", 1, opcodeRoll},
	OP_ROT:          {OP_ROT, "OP_ROT", 1, opcodeRot},
	OP_SWAP:         {OP_SWAP, "OP_SWAP", 1, opcodeSwap},
	OP_TUCK:         {OP_TUCK, "OP_TUCK", 1, opcodeTuck},

	// Splice opcodes.
	OP_CAT:    {OP_CAT, "OP_CAT", 1, opcodeDisabled},
	OP_SUBSTR: {OP_SUBSTR, "OP_SUBSTR", 1, opcodeDisabled},
	OP_LEFT:   {OP_LEFT, "OP_LEFT", 1, opcodeDisabled},
	OP_RIGHT:  {OP_RIGHT, "OP_RIGHT", 1, opcodeDisabled},
	OP_SIZE:   {OP_SIZE, "OP_SIZE", 1, opcodeSize},

	// Bitwise logic opcodes.
	OP_INVERT:      {OP_INVERT, "OP_INVERT", 1, opcodeDisabled},
	OP_AND:         {OP_AND, "OP_AND", 1, opcodeDisabled},
	OP_OR:          {OP_OR, "OP_OR", 1, opcodeDisabled},
	OP_XOR:         {OP_XOR, "OP_XOR", 1, opcodeDisabled},
	OP_EQUAL:       {OP_EQUAL, "OP_EQUAL", 1, opcodeEqual},
	OP_EQUALVERIFY: {OP_EQUALVERIFY, "OP_EQUALVERIFY", 1, opcodeEqualVerify},
	OP_RESERVED1:   {OP_RESERVED1, "OP_RESERVED1", 1, opcodeReserved},
	OP_RESERVED2:   {OP_RESERVED2, "OP_RESERVED2", 1, opcodeReserved},

	// Numeric related opcodes.
	OP_1ADD:               {OP_1ADD, "OP_1ADD", 1, opcode1Add},
	OP_1SUB:               {OP_1SUB, "OP_1SUB", 1, opcode1Sub},
	OP_2MUL:               {OP_2MUL, "OP_2MUL", 1, opcodeDisabled},
	OP_2DIV:               {OP_2DIV, "OP_2DIV", 1, opcodeDisabled},
	OP_NEGATE:             {OP_NEGATE, "OP_NEGATE", 1, opcodeNegate},
	OP_ABS:                {OP_ABS, "OP_ABS", 1, opcodeAbs},
	OP_NOT:                {OP_NOT, "OP_NOT", 1, opcodeNot},
	OP_0NOTEQUAL:          {OP_0NOTEQUAL, "OP_0NOTEQUAL", 1, opcode0NotEqual},
	OP_ADD:                {OP_ADD, "OP_ADD", 1, opcodeAdd},
	OP_SUB:                {OP_SUB, "OP_SUB", 1, opcodeSub},
	OP_MUL:                {OP_MUL, "OP_MUL", 1, opcodeDisabled},
	OP_DIV:                {OP_DIV, "OP_DIV", 1, opcodeDisabled},
	OP_MOD:                {OP_MOD, "OP_MOD", 1, opcodeDisabled},
	OP_LSHIFT:             {OP_LSHIFT, "OP_LSHIFT", 1, opcodeDisabled},
	OP_RSHIFT:             {OP_RSHIFT, "OP_RSHIFT", 1, opcodeDisabled},
	OP_BOOLAND:            {OP_BOOLAND, "OP_BOOLAND", 1, opcodeBoolAnd},
	OP_BOOLOR:             {OP_BOOLOR, "OP_BOOLOR", 1, opcodeBoolOr},
	OP_NUMEQUAL:           {OP_NUMEQUAL, "OP_NUMEQUAL", 1, opcodeNumEqual},
	OP_NUMEQUALVERIFY:     {OP_NUMEQUALVERIFY, "OP_NUMEQUALVERIFY", 1, opcodeNumEqualVerify},
	OP_NUMNOTEQUAL:        {OP_NUMNOTEQUAL, "OP_NUMNOTEQUAL", 1, opcodeNumNotEqual},
	OP_LESSTHAN:           {OP_LESSTHAN, "OP_LESSTHAN", 1, opcodeLessThan},
	OP_GREATERTHAN:        {OP_GREATERTHAN, "OP_GREATERTHAN", 1, opcodeGreaterThan},
	OP_LESSTHANOREQUAL:    {OP_LESSTHANOREQUAL, "OP_LESSTHANOREQUAL", 1, opcodeLessThanOrEqual},
	OP_GREATERTHANOREQUAL: {OP_GREATERTHANOREQUAL, "OP_GREATERTHANOREQUAL", 1, opcodeGreaterThanOrEqual},
	OP_MIN:                {OP_MIN, "OP_MIN", 1, opcodeMin},
	OP_MAX:                {OP_MAX, "OP_MAX", 1, opcodeMax},
	OP_WITHIN:             {OP_WITHIN, "OP_WITHIN", 1, opcodeWithin},

	// Crypto opcodes.
	OP_RIPEMD160:           {OP_RIPEMD160, "OP_RIPEMD160", 1, opcodeRipemd160},
	OP_SHA1:                {OP_SHA1, "OP_SHA1", 1, opcodeSha1},
	OP_SHA256:              {OP_SHA256, "OP_SHA256", 1, opcodeSha256},
	OP_HASH160:             {OP_HASH160, "OP_HASH160", 1, opcodeHash160},
	OP_HASH256:             {OP_HASH256, "OP_HASH256", 1, opcodeHash256},
	OP_CODESEPARATOR:       {OP_CODESEPARATOR, "OP_CODESEPARATOR", 1, opcodeCodeSeparator},
	OP_CHECKSIG:            {OP_CHECKSIG, "OP_CHECKSIG", 1, opcodeCheckSig},
	OP_CHECKSIGVERIFY:      {OP_CHECKSIGVERIFY, "OP_CHECKSIGVERIFY", 1, opcodeCheckSigVerify},
	OP_CHECKMULTISIG:       {OP_CHECKMULTISIG, "OP_CHECKMULTISIG", 1, opcodeCheckMultiSig},
	OP_CHECKMULTISIGVERIFY: {OP_CHECKMULTISIGVERIFY, "OP_CHECKMULTISIGVERIFY", 1, opcodeCheckMultiSigVerify},

	// Reserved opcodes.
	OP_NOP1:  {OP_NOP1, "OP_NOP1", 1, opcodeNop},
	OP_NOP2:  {OP_NOP2, "OP_NOP2", 1, opcodeNop},
	OP_NOP3:  {OP_NOP3, "OP_NOP3", 1, opcodeNop},
	OP_NOP4:  {OP_NOP4, "OP_NOP4", 1, opcodeNop},
	OP_NOP5:  {OP_NOP5, "OP_NOP5", 1, opcodeNop},
	OP_NOP6:  {OP_NOP6, "OP_NOP6", 1, opcodeNop},
	OP_NOP7:  {OP_NOP7, "OP_NOP7", 1, opcodeNop},
	OP_NOP8:  {OP_NOP8, "OP_NOP8", 1, opcodeNop},
	OP_NOP9:  {OP_NOP9, "OP_NOP9", 1, opcodeNop},
	OP_NOP10: {OP_NOP10, "OP_NOP10", 1, opcodeNop},

	// Undefined opcodes.
	OP_UNKNOWN186: {OP_UNKNOWN186, "OP_UNKNOWN186", 1, opcodeInvalid},
	OP_UNKNOWN187: {OP_UNKNOWN187, "OP_UNKNOWN187", 1, opcodeInvalid},
	OP_UNKNOWN188: {OP_UNKNOWN188, "OP_UNKNOWN188", 1, opcodeInvalid},
	OP_UNKNOWN189: {OP_UNKNOWN189, "OP_UNKNOWN189", 1, opcodeInvalid},
	OP_UNKNOWN190: {OP_UNKNOWN190, "OP_UNKNOWN190", 1, opcodeInvalid},
	OP_UNKNOWN191: {OP_UNKNOWN191, "OP_UNKNOWN191", 1, opcodeInvalid},
	OP_UNKNOWN192: {OP_UNKNOWN192, "OP_UNKNOWN192", 1, opcodeInvalid},
	OP_UNKNOWN193: {OP_UNKNOWN193, "OP_UNKNOWN193", 1, opcodeInvalid},
	OP_UNKNOWN194: {OP_UNKNOWN194, "OP_UNKNOWN194", 1, opcodeInvalid},
	OP_UNKNOWN195: {OP_UNKNOWN195, "OP_UNKNOWN195", 1, opcodeInvalid},
	OP_UNKNOWN196: {OP_UNKNOWN196, "OP_UNKNOWN196", 1, opcodeInvalid},
	OP_UNKNOWN197: {OP_UNKNOWN197, "OP_UNKNOWN197", 1, opcodeInvalid},
	OP_UNKNOWN198: {OP_UNKNOWN198, "OP_UNKNOWN198", 1, opcodeInvalid},
	OP_UNKNOWN199: {OP_UNKNOWN199, "OP_UNKNOWN199", 1, opcodeInvalid},
	OP_UNKNOWN200: {OP_UNKNOWN200, "OP_UNKNOWN200", 1, opcodeInvalid},
	OP_UNKNOWN201: {OP_UNKNOWN201, "OP_UNKNOWN201", 1, opcodeInvalid},
	OP_UNKNOWN202: {OP_UNKNOWN202, "OP_UNKNOWN202", 1, opcodeInvalid},
	OP_UNKNOWN203: {OP_UNKNOWN203, "OP_UNKNOWN203", 1, opcodeInvalid},
	OP_UNKNOWN204: {OP_UNKNOWN204, "OP_UNKNOWN204", 1, opcodeInvalid},
	OP_UNKNOWN205: {OP_UNKNOWN205, "OP_UNKNOWN205", 1, opcodeInvalid},
	OP_UNKNOWN206: {OP_UNKNOWN206, "OP_UNKNOWN206", 1, opcodeInvalid},
	OP_UNKNOWN207: {OP_UNKNOWN207, "OP_UNKNOWN207", 1, opcodeInvalid},
	OP_UNKNOWN208: {OP_UNKNOWN208, "OP_UNKNOWN208", 1, opcodeInvalid},
	OP_UNKNOWN209: {OP_UNKNOWN209, "OP_UNKNOWN209", 1, opcodeInvalid},
	OP_UNKNOWN210: {OP_UNKNOWN210, "OP_UNKNOWN210", 1, opcodeInvalid},
	OP_UNKNOWN211: {OP_UNKNOWN211, "OP_UNKNOWN211", 1, opcodeInvalid},
	OP_UNKNOWN212: {OP_UNKNOWN212, "OP_UNKNOWN212", 1, opcodeInvalid},
	OP_UNKNOWN213: {OP_UNKNOWN213, "OP_UNKNOWN213", 1, opcodeInvalid},
	OP_UNKNOWN214: {OP_UNKNOWN214, "OP_UNKNOWN214", 1, opcodeInvalid},
	OP_UNKNOWN215: {OP_UNKNOWN215, "OP_UNKNOWN215", 1, opcodeInvalid},
	OP_UNKNOWN216: {OP_UNKNOWN216, "OP_UNKNOWN216", 1, opcodeInvalid},
	OP_UNKNOWN217: {OP_UNKNOWN217, "OP_UNKNOWN217", 1, opcodeInvalid},
	OP_UNKNOWN218: {OP_UNKNOWN218, "OP_UNKNOWN218", 1, opcodeInvalid},
	OP_UNKNOWN219: {OP_UNKNOWN219, "OP_UNKNOWN219", 1, opcodeInvalid},
	OP_UNKNOWN220: {OP_UNKNOWN220, "OP_UNKNOWN220", 1, opcodeInvalid},
	OP_UNKNOWN221: {OP_UNKNOWN221, "OP_UNKNOWN221", 1, opcodeInvalid},
	OP_UNKNOWN222: {OP_UNKNOWN222, "OP_UNKNOWN222", 1, opcodeInvalid},
	OP_UNKNOWN223: {OP_UNKNOWN223, "OP_UNKNOWN223", 1, opcodeInvalid},
	OP_UNKNOWN224: {OP_UNKNOWN224, "OP_UNKNOWN224", 1, opcodeInvalid},
	OP_UNKNOWN225: {OP_UNKNOWN225, "OP_UNKNOWN225", 1, opcodeInvalid},
	OP_UNKNOWN226: {OP_UNKNOWN226, "OP_UNKNOWN226", 1, opcodeInvalid},
	OP_UNKNOWN227: {OP_UNKNOWN227, "OP_UNKNOWN227", 1, opcodeInvalid},
	OP_UNKNOWN228: {OP_UNKNOWN228, "OP_UNKNOWN228", 1, opcodeInvalid},
	OP_UNKNOWN229: {OP_UNKNOWN229, "OP_UNKNOWN229", 1, opcodeInvalid},
	OP_UNKNOWN230: {OP_UNKNOWN230, "OP_UNKNOWN230", 1, opcodeInvalid},
	OP_UNKNOWN231: {OP_UNKNOWN231, "OP_UNKNOWN231", 1, opcodeInvalid},
	OP_UNKNOWN232: {OP_UNKNOWN232, "OP_UNKNOWN232", 1, opcodeInvalid},
	OP_UNKNOWN233: {OP_UNKNOWN233, "OP_UNKNOWN233", 1, opcodeInvalid},
	OP_UNKNOWN234: {OP_UNKNOWN234, "OP_UNKNOWN234", 1, opcodeInvalid},
	OP_UNKNOWN235: {OP_UNKNOWN235, "OP_UNKNOWN235", 1, opcodeInvalid},
	OP_UNKNOWN236: {OP_UNKNOWN236, "OP_UNKNOWN236", 1, opcodeInvalid},
	OP_UNKNOWN237: {OP_UNKNOWN237, "OP_UNKNOWN237", 1, opcodeInvalid},
	OP_UNKNOWN238: {OP_UNKNOWN238, "OP_UNKNOWN238", 1, opcodeInvalid},
	OP_UNKNOWN239: {OP_UNKNOWN239, "OP_UNKNOWN239", 1, opcodeInvalid},
	OP_UNKNOWN240: {OP_UNKNOWN240, "OP_UNKNOWN240", 1, opcodeInvalid},
	OP_UNKNOWN241: {OP_UNKNOWN241, "OP_UNKNOWN241", 1, opcodeInvalid},
	OP_UNKNOWN242: {OP_UNKNOWN242, "OP_UNKNOWN242", 1, opcodeInvalid},
	OP_UNKNOWN243: {OP_UNKNOWN243, "OP_UNKNOWN243", 1, opcodeInvalid},
	OP_UNKNOWN244: {OP_UNKNOWN244, "OP_UNKNOWN244", 1, opcodeInvalid},
	OP_UNKNOWN245: {OP_UNKNOWN245, "OP_UNKNOWN245", 1, opcodeInvalid},
	OP_UNKNOWN246: {OP_UNKNOWN246, "OP_UNKNOWN246", 1, opcodeInvalid},
	OP_UNKNOWN247: {OP_UNKNOWN247, "OP_UNKNOWN247", 1, opcodeInvalid},
	OP_UNKNOWN248: {OP_UNKNOWN248, "OP_UNKNOWN248", 1, opcodeInvalid},

	// Bitcoin Core internal use opcode.  Defined here for completeness.
	OP_SMALLDATA:    {OP_SMALLDATA, "OP_SMALLDATA", 1, opcodeInvalid},
	OP_SMALLINTEGER: {OP_SMALLINTEGER, "OP_SMALLINTEGER", 1, opcodeInvalid},
	OP_PUBKEYS:      {OP_PUBKEYS, "OP_PUBKEYS", 1, opcodeInvalid},
	OP_UNKNOWN252:   {OP_UNKNOWN252, "OP_UNKNOWN252", 1, opcodeInvalid},
	OP_PUBKEYHASH:   {OP_PUBKEYHASH, "OP_PUBKEYHASH", 1, opcodeInvalid},
	OP_PUBKEY:       {OP_PUBKEY, "OP_PUBKEY", 1, opcodeInvalid},

	OP_INVALIDOPCODE: {OP_INVALIDOPCODE, "OP_INVALIDOPCODE", 1, opcodeInvalid},
}

// opcodeOnelineRepls defines opcode names which are replaced when doing a
// one-line disassembly.  This is done to match the output of the reference
// implementation while not changing the opcode names in the nicer full
// disassembly.
var opcodeOnelineRepls = map[string]string{
	"OP_1NEGATE": "-1",
	"OP_0":       "0",
	"OP_1":       "1",
	"OP_2":       "2",
	"OP_3":       "3",
	"OP_4":       "4",
	"OP_5":       "5",
	"OP_6":       "6",
	"OP_7":       "7",
	"OP_8":       "8",
	"OP_9":       "9",
	"OP_10":      "10",
	"OP_11":      "11",
	"OP_12":      "12",
	"OP_13":      "13",
	"OP_14":      "14",
	"OP_15":      "15",
	"OP_16":      "16",
}

type parsedOpcode struct {
	opcode *opcode
	data   []byte
}

// The following opcodes are disabled and are thus always bad to see in the
// instruction stream (even if turned off by a conditional).
func (pop *parsedOpcode) disabled() bool {
	switch pop.opcode.value {
	case OP_CAT:
		return true
	case OP_SUBSTR:
		return true
	case OP_LEFT:
		return true
	case OP_RIGHT:
		return true
	case OP_INVERT:
		return true
	case OP_AND:
		return true
	case OP_OR:
		return true
	case OP_XOR:
		return true
	case OP_2MUL:
		return true
	case OP_2DIV:
		return true
	case OP_MUL:
		return true
	case OP_DIV:
		return true
	case OP_MOD:
		return true
	case OP_LSHIFT:
		return true
	case OP_RSHIFT:
		return true
	default:
		return false
	}
}

// The following opcodes are always illegal when passed over by the program
// counter even if in a non-executed branch. (it isn't a coincidence that they
// are conditionals).
func (pop *parsedOpcode) alwaysIllegal() bool {
	switch pop.opcode.value {
	case OP_VERIF:
		return true
	case OP_VERNOTIF:
		return true
	default:
		return false
	}
}

// The following opcode are conditional and thus change the conditional
// execution stack state when passed.
func (pop *parsedOpcode) conditional() bool {
	switch pop.opcode.value {
	case OP_IF:
		return true
	case OP_NOTIF:
		return true
	case OP_ELSE:
		return true
	case OP_ENDIF:
		return true
	default:
		return false
	}
}

// checkMinimalDataPush returns whether or not the current data
// push uses the correct opcode.
func (pop *parsedOpcode) checkMinimalDataPush() error {
	data := pop.data
	dataLen := len(data)
	opcode := pop.opcode.value

	if dataLen == 0 && opcode != OP_0 {
		return ErrStackMinimalData
	} else if dataLen == 1 && data[0] >= 1 && data[0] <= 16 {
		if opcode != OP_1+data[0]-1 {
			// Should have used OP_1 .. OP_16
			return ErrStackMinimalData
		}
	} else if dataLen == 1 && data[0] == 0x81 {
		if opcode != OP_1NEGATE {
			return ErrStackMinimalData
		}
	} else if dataLen <= 75 {
		if int(opcode) != dataLen {
			// Should have used a direct push
			return ErrStackMinimalData
		}
	} else if dataLen <= 255 {
		if opcode != OP_PUSHDATA1 {
			return ErrStackMinimalData
		}
	} else if dataLen <= 65535 {
		if opcode != OP_PUSHDATA2 {
			return ErrStackMinimalData
		}
	}
	return nil
}

// exec peforms execution on the opcode. It takes into account whether or not
// it is hidden by conditionals, but some rules still must be tested in this
// case.
func (pop *parsedOpcode) exec(vm *Engine) error {
	// Disabled opcodes are ``fail on program counter''.
	if pop.disabled() {
		return ErrStackOpDisabled
	}

	// Always-illegal opcodes are ``fail on program counter''.
	if pop.alwaysIllegal() {
		return ErrStackReservedOpcode
	}

	// Note that this includes OP_RESERVED which counts as a push operation.
	if pop.opcode.value > OP_16 {
		vm.numOps++
		if vm.numOps > MaxOpsPerScript {
			return ErrStackTooManyOperations
		}

	} else if len(pop.data) > MaxScriptElementSize {
		return ErrStackElementTooBig
	}

	// If we are not a conditional opcode and we aren't executing, then
	// we are done now.
	if vm.condStack[0] != OpCondTrue && !pop.conditional() {
		return nil
	}

	// Ensure all executed data push opcodes use the minimal encoding when
	// the minimal data verification is set.
	if vm.dstack.verifyMinimalData && vm.condStack[0] == OpCondTrue &&
		pop.opcode.value >= 0 && pop.opcode.value <= OP_PUSHDATA4 {
		if err := pop.checkMinimalDataPush(); err != nil {
			return err
		}
	}

	return pop.opcode.opfunc(pop, vm)
}

func (pop *parsedOpcode) print(oneline bool) string {
	// The reference implementation one-line disassembly replaces opcodes
	// which represent values (e.g. OP_0 through OP_16 and OP_1NEGATE)
	// with the raw value.  However, when not doing a one-line dissassembly,
	// we prefer to show the actual opcode names.  Thus, only replace the
	// opcodes in question when the oneline flag is set.
	opcodeName := pop.opcode.name
	if oneline {
		if replName, ok := opcodeOnelineRepls[opcodeName]; ok {
			opcodeName = replName
		}
	}

	retString := opcodeName
	if pop.opcode.length == 1 {
		return retString
	}
	if oneline {
		retString = ""
	}
	if !oneline && pop.opcode.length < 0 {
		//add length to the end of retString
		retString += fmt.Sprintf(" 0x%0*x", 2*-pop.opcode.length,
			len(pop.data))
	}
	for _, val := range pop.data {
		if !oneline {
			retString += " "
		}
		retString += fmt.Sprintf("%02x", val)
	}
	return retString
}

func (pop *parsedOpcode) bytes() ([]byte, error) {
	var retbytes []byte
	if pop.opcode.length > 0 {
		retbytes = make([]byte, 1, pop.opcode.length)
	} else {
		retbytes = make([]byte, 1, 1+len(pop.data)-
			pop.opcode.length)
	}

	retbytes[0] = pop.opcode.value
	if pop.opcode.length == 1 {
		if len(pop.data) != 0 {
			return nil, ErrStackInvalidOpcode
		}
		return retbytes, nil
	}
	nbytes := pop.opcode.length
	if pop.opcode.length < 0 {
		l := len(pop.data)
		// tempting just to hardcode to avoid the complexity here.
		switch pop.opcode.length {
		case -1:
			retbytes = append(retbytes, byte(l))
			nbytes = int(retbytes[1]) + len(retbytes)
		case -2:
			retbytes = append(retbytes, byte(l&0xff),
				byte(l>>8&0xff))
			nbytes = int(binary.LittleEndian.Uint16(retbytes[1:])) +
				len(retbytes)
		case -4:
			retbytes = append(retbytes, byte(l&0xff),
				byte((l>>8)&0xff), byte((l>>16)&0xff),
				byte((l>>24)&0xff))
			nbytes = int(binary.LittleEndian.Uint32(retbytes[1:])) +
				len(retbytes)
		}
	}

	retbytes = append(retbytes, pop.data...)

	if len(retbytes) != nbytes {
		return nil, ErrStackInvalidOpcode
	}

	return retbytes, nil
}

// opcode implementation functions from here

func opcodeDisabled(op *parsedOpcode, vm *Engine) error {
	return ErrStackOpDisabled
}

func opcodeReserved(op *parsedOpcode, vm *Engine) error {
	return ErrStackReservedOpcode
}

// Recognised opcode, but for bitcoind internal use only.
func opcodeInvalid(op *parsedOpcode, vm *Engine) error {
	return ErrStackInvalidOpcode
}

func opcodeFalse(op *parsedOpcode, vm *Engine) error {
	vm.dstack.PushByteArray([]byte(""))

	return nil
}

func opcodePushData(op *parsedOpcode, vm *Engine) error {
	vm.dstack.PushByteArray(op.data)
	return nil
}

func opcode1Negate(op *parsedOpcode, vm *Engine) error {
	vm.dstack.PushInt(big.NewInt(-1))
	return nil
}

func opcodeN(op *parsedOpcode, vm *Engine) error {
	// 16 consecutive opcodes add increasing numbers to the stack.
	vm.dstack.PushInt(big.NewInt(int64(op.opcode.value - (OP_1 - 1))))
	return nil
}

func opcodeNop(op *parsedOpcode, vm *Engine) error {
	switch op.opcode.value {
	case OP_NOP1, OP_NOP2, OP_NOP3, OP_NOP4, OP_NOP5,
		OP_NOP6, OP_NOP7, OP_NOP8, OP_NOP9, OP_NOP10:
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return fmt.Errorf("OP_NOP%d reserved for soft-fork "+
				"upgrades", op.opcode.value-OP_NOP1+1)
		}
	}
	return nil
}

// opcodeIf computes true/false based on the value on the stack and pushes
// the condition on the condStack (conditional execution stack)
func opcodeIf(op *parsedOpcode, vm *Engine) error {
	// opcodeIf will be executed even if it is on the non-execute side
	// of the conditional, this is so proper nesting is maintained
	var condval int
	if vm.condStack[0] == OpCondTrue {
		ok, err := vm.dstack.PopBool()
		if err != nil {
			return err
		}
		if ok {
			condval = OpCondTrue
		}
	} else {
		condval = OpCondSkip
	}
	cond := []int{condval}
	// push condition to the 'head' of the slice
	vm.condStack = append(cond, vm.condStack...)
	// TODO(drahn) check if a maximum condtitional stack limit exists
	return nil
}

// opcodeNotIf computes true/false based on the value on the stack and pushes
// the (inverted) condition on the condStack (conditional execution stack)
func opcodeNotIf(op *parsedOpcode, vm *Engine) error {
	// opcodeIf will be executed even if it is on the non-execute side
	// of the conditional, this is so proper nesting is maintained
	var condval int
	if vm.condStack[0] == OpCondTrue {
		ok, err := vm.dstack.PopBool()
		if err != nil {
			return err
		}
		if !ok {
			condval = OpCondTrue
		}
	} else {
		condval = OpCondSkip
	}
	cond := []int{condval}
	// push condition to the 'head' of the slice
	vm.condStack = append(cond, vm.condStack...)
	// TODO(drahn) check if a maximum condtitional stack limit exists
	return nil
}

// opcodeElse inverts conditional execution for other half of if/else/endif
func opcodeElse(op *parsedOpcode, vm *Engine) error {
	if len(vm.condStack) < 2 {
		// intial true cannot be toggled, only pushed conditionals
		return ErrStackNoIf
	}

	switch vm.condStack[0] {
	case OpCondTrue:
		vm.condStack[0] = OpCondFalse
	case OpCondFalse:
		vm.condStack[0] = OpCondTrue
	case OpCondSkip:
		// value doesn't change in skip
	}
	return nil
}

// opcodeEndif terminates a conditional block, removing the  value from the
// conditional execution stack.
func opcodeEndif(op *parsedOpcode, vm *Engine) error {
	if len(vm.condStack) < 2 {
		// intial true cannot be popped, only pushed conditionals
		return ErrStackNoIf
	}

	stk := make([]int, len(vm.condStack)-1, len(vm.condStack)-1)
	copy(stk, vm.condStack[1:])
	vm.condStack = stk
	return nil
}

func opcodeVerify(op *parsedOpcode, vm *Engine) error {
	verified, err := vm.dstack.PopBool()
	if err != nil {
		return err
	}

	if verified != true {
		return ErrStackVerifyFailed
	}
	return nil
}

func opcodeReturn(op *parsedOpcode, vm *Engine) error {
	return ErrStackEarlyReturn
}

func opcodeToAltStack(op *parsedOpcode, vm *Engine) error {
	so, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}
	vm.astack.PushByteArray(so)

	return nil
}

func opcodeFromAltStack(op *parsedOpcode, vm *Engine) error {
	so, err := vm.astack.PopByteArray()
	if err != nil {
		return err
	}
	vm.dstack.PushByteArray(so)

	return nil
}

func opcode2Drop(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.DropN(2)
}

func opcode2Dup(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.DupN(2)
}

func opcode3Dup(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.DupN(3)
}

func opcode2Over(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.OverN(2)
}

func opcode2Rot(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.RotN(2)
}

func opcode2Swap(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.SwapN(2)
}

func opcodeIfDup(op *parsedOpcode, vm *Engine) error {
	val, err := vm.dstack.PeekInt(0)
	if err != nil {
		return err
	}

	// Push copy of data iff it isn't zero
	if val.Sign() != 0 {
		vm.dstack.PushInt(val)
	}

	return nil
}

func opcodeDepth(op *parsedOpcode, vm *Engine) error {
	vm.dstack.PushInt(big.NewInt(int64(vm.dstack.Depth())))
	return nil
}

func opcodeDrop(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.DropN(1)
}

func opcodeDup(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.DupN(1)
}

func opcodeNip(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.NipN(1)
}

func opcodeOver(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.OverN(1)
}

// Copy object N items back in the stack to the top. Where N is the value in
// the top of the stack.
func opcodePick(op *parsedOpcode, vm *Engine) error {
	pidx, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	// PopInt promises that the int returned is 32 bit.
	val := int(pidx.Int64())

	return vm.dstack.PickN(val)
}

// Move object N items back in the stack to the top. Where N is the value in
// the top of the stack.
func opcodeRoll(op *parsedOpcode, vm *Engine) error {
	ridx, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	// PopInt promises that the int returned is 32 bit.
	val := int(ridx.Int64())

	return vm.dstack.RollN(val)
}

// Rotate top three items on the stack to the left.
// e.g. 1,2,3 -> 2,3,1
func opcodeRot(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.RotN(1)
}

// Swap the top two items on the stack: 1,2 -> 2,1
func opcodeSwap(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.SwapN(1)
}

// The item at the top of the stack is copied and inserted before the
// second-to-top item. e.g.: 2,1, -> 2,1,2
func opcodeTuck(op *parsedOpcode, vm *Engine) error {
	return vm.dstack.Tuck()
}

// Push the size of the item on top of the stack onto the stack.
func opcodeSize(op *parsedOpcode, vm *Engine) error {
	i, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	vm.dstack.PushInt(big.NewInt(int64(len(i))))
	return nil
}

func opcodeEqual(op *parsedOpcode, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}
	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushBool(bytes.Equal(a, b))
	return nil
}

func opcodeEqualVerify(op *parsedOpcode, vm *Engine) error {
	err := opcodeEqual(op, vm)
	if err == nil {
		err = opcodeVerify(op, vm)
	}
	return err
}

func opcode1Add(op *parsedOpcode, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(new(big.Int).Add(m, big.NewInt(1)))

	return nil
}

func opcode1Sub(op *parsedOpcode, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	vm.dstack.PushInt(new(big.Int).Sub(m, big.NewInt(1)))

	return nil
}

func opcodeNegate(op *parsedOpcode, vm *Engine) error {
	// XXX when we remove types just flip the 0x80 bit of msb
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(new(big.Int).Neg(m))
	return nil
}

func opcodeAbs(op *parsedOpcode, vm *Engine) error {
	// XXX when we remove types just &= ~0x80 on msb
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(new(big.Int).Abs(m))

	return nil
}

// If then input is 0 or 1, it is flipped. Otherwise the output will be 0.
// (n.b. official client just has 1 is 0, else 0)
func opcodeNot(op *parsedOpcode, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	if m.Sign() == 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}
	return nil
}

// opcode returns 0 if the input is 0, 1 otherwise.
func opcode0NotEqual(op *parsedOpcode, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	if m.Sign() != 0 {
		m.SetInt64(1)
	}
	vm.dstack.PushInt(m)

	return nil
}

// Push result of adding top two entries on stack
func opcodeAdd(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(new(big.Int).Add(v0, v1))
	return nil
}

// Push result of subtracting 2nd entry on stack from first.
func opcodeSub(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(new(big.Int).Sub(v1, v0))
	return nil
}

// If both of the top two entries on the stack are not zero output is 1.
// Otherwise, 0.
func opcodeBoolAnd(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0.Sign() != 0 && v1.Sign() != 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

// If either of the top two entries on the stack are not zero output is 1.
// Otherwise, 0.
func opcodeBoolOr(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0.Sign() != 0 || v1.Sign() != 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

func opcodeNumEqual(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0.Cmp(v1) == 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

func opcodeNumEqualVerify(op *parsedOpcode, vm *Engine) error {
	err := opcodeNumEqual(op, vm)
	if err == nil {
		err = opcodeVerify(op, vm)
	}
	return err
}

func opcodeNumNotEqual(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0.Cmp(v1) != 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

func opcodeLessThan(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) == -1 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

func opcodeGreaterThan(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) == 1 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}
	return nil
}

func opcodeLessThanOrEqual(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) <= 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}
	return nil
}

func opcodeGreaterThanOrEqual(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) >= 0 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}

	return nil
}

func opcodeMin(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) == -1 {
		vm.dstack.PushInt(new(big.Int).Set(v1))
	} else {
		vm.dstack.PushInt(new(big.Int).Set(v0))
	}

	return nil
}

func opcodeMax(op *parsedOpcode, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1.Cmp(v0) == 1 {
		vm.dstack.PushInt(new(big.Int).Set(v1))
	} else {
		vm.dstack.PushInt(new(big.Int).Set(v0))
	}

	return nil
}

// stack input: x, min, max. Returns 1 if x is within specified range
// (left inclusive), 0 otherwise
func opcodeWithin(op *parsedOpcode, vm *Engine) error {
	maxVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	minVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	x, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if x.Cmp(minVal) >= 0 && x.Cmp(maxVal) == -1 {
		vm.dstack.PushInt(big.NewInt(1))
	} else {
		vm.dstack.PushInt(big.NewInt(0))
	}
	return nil
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// calculate hash160 which is ripemd160(sha256(data))
func calcHash160(buf []byte) []byte {
	return calcHash(calcHash(buf, fastsha256.New()), ripemd160.New())
}

func opcodeRipemd160(op *parsedOpcode, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(calcHash(buf, ripemd160.New()))
	return nil
}

func opcodeSha1(op *parsedOpcode, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(calcHash(buf, sha1.New()))
	return nil
}

func opcodeSha256(op *parsedOpcode, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(calcHash(buf, fastsha256.New()))
	return nil
}

func opcodeHash160(op *parsedOpcode, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(calcHash160(buf))
	return nil
}

func opcodeHash256(op *parsedOpcode, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(wire.DoubleSha256(buf))
	return nil
}

func opcodeCodeSeparator(op *parsedOpcode, vm *Engine) error {
	vm.lastcodesep = vm.scriptOff

	return nil
}

func opcodeCheckSig(op *parsedOpcode, vm *Engine) error {

	pkStr, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	sigStr, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// Signature actually needs needs to be longer than this, but we need
	// at least  1 byte for the below. btcec will check full length upon
	// parsing the signature.
	if len(sigStr) < 1 {
		vm.dstack.PushBool(false)
		return nil
	}

	// Trim off hashtype from the signature string.
	hashType := SigHashType(sigStr[len(sigStr)-1])
	sigStr = sigStr[:len(sigStr)-1]

	if err := vm.checkHashTypeEncoding(hashType); err != nil {
		return err
	}
	if err := vm.checkSignatureEncoding(sigStr); err != nil {
		return err
	}
	if err := vm.checkPubKeyEncoding(pkStr); err != nil {
		return err
	}

	// Get script from the last OP_CODESEPARATOR and without any subsequent
	// OP_CODESEPARATORs
	subScript := vm.subScript()

	// Unlikely to hit any cases here, but remove the signature from
	// the script if present.
	subScript = removeOpcodeByData(subScript, sigStr)

	hash := calcScriptHash(subScript, hashType, &vm.tx, vm.txIdx)

	pubKey, err := btcec.ParsePubKey(pkStr, btcec.S256())
	if err != nil {
		vm.dstack.PushBool(false)
		return nil
	}

	var signature *btcec.Signature
	if vm.hasFlag(ScriptVerifyStrictEncoding) ||
		vm.hasFlag(ScriptVerifyDERSignatures) {

		signature, err = btcec.ParseDERSignature(sigStr, btcec.S256())
	} else {
		signature, err = btcec.ParseSignature(sigStr, btcec.S256())
	}
	if err != nil {
		vm.dstack.PushBool(false)
		return nil
	}

	log.Tracef("%v", newLogClosure(func() string {
		return fmt.Sprintf("op_checksig\n"+
			"pubKey:\n%v"+
			"pubKey.X: %v\n"+
			"pubKey.Y: %v\n"+
			"signature.R: %v\n"+
			"signature.S: %v\n"+
			"checkScriptHash:\n%v",
			hex.Dump(pkStr), pubKey.X, pubKey.Y,
			signature.R, signature.S, hex.Dump(hash))
	}))
	ok := signature.Verify(hash, pubKey)
	vm.dstack.PushBool(ok)
	return nil
}

func opcodeCheckSigVerify(op *parsedOpcode, vm *Engine) error {
	err := opcodeCheckSig(op, vm)
	if err == nil {
		err = opcodeVerify(op, vm)
	}
	return err
}

// parsedSigInfo houses a raw signature along with its parsed form and a flag
// for whether or not it has already been parsed.  It is used to prevent parsing
// the same signature multiple times when verify a multisig.
type parsedSigInfo struct {
	signature       []byte
	parsedSignature *btcec.Signature
	parsed          bool
}

// stack; sigs <numsigs> pubkeys <numpubkeys>
func opcodeCheckMultiSig(op *parsedOpcode, vm *Engine) error {
	numKeys, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	// PopInt promises that the int returned is 32 bit.
	numPubKeys := int(numKeys.Int64())
	if numPubKeys < 0 || numPubKeys > MaxPubKeysPerMultiSig {
		return ErrStackTooManyPubkeys
	}
	vm.numOps += numPubKeys
	if vm.numOps > MaxOpsPerScript {
		return ErrStackTooManyOperations
	}

	pubKeys := make([][]byte, 0, numPubKeys)
	for i := 0; i < numPubKeys; i++ {
		pubKey, err := vm.dstack.PopByteArray()
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pubKey)
	}

	numSigs, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	// PopInt promises that the int returned is 32 bit.
	numSignatures := int(numSigs.Int64())
	if numSignatures < 0 {
		return fmt.Errorf("number of signatures '%d' is less than 0",
			numSignatures)
	}
	if numSignatures > numPubKeys {
		return fmt.Errorf("more signatures than pubkeys: %d > %d",
			numSignatures, numPubKeys)
	}

	signatures := make([]*parsedSigInfo, 0, numSignatures)
	for i := 0; i < numSignatures; i++ {
		signature, err := vm.dstack.PopByteArray()
		if err != nil {
			return err
		}
		sigInfo := &parsedSigInfo{signature: signature}
		signatures = append(signatures, sigInfo)
	}

	// bug in bitcoind means we pop one more stack value than should be
	// used.
	dummy, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if vm.hasFlag(ScriptStrictMultiSig) && len(dummy) != 0 {
		return fmt.Errorf("multisig dummy argument is not zero length: %d",
			len(dummy))
	}

	// Trim OP_CODESEPARATORs
	script := vm.subScript()

	// Remove any of the signatures that happen to be in the script.
	// can't sign somthing containing the signature you're making, after
	// all
	for _, sigInfo := range signatures {
		script = removeOpcodeByData(script, sigInfo.signature)
	}

	success := true
	numPubKeys++
	pubKeyIdx := -1
	signatureIdx := 0
	for numSignatures > 0 {
		// When there are more signatures than public keys remaining,
		// there is no way to succeed since too many signatures are
		// invalid, so exit early.
		pubKeyIdx++
		numPubKeys--
		if numSignatures > numPubKeys {
			success = false
			break
		}

		sigInfo := signatures[signatureIdx]
		pubKey := pubKeys[pubKeyIdx]

		// The order of the signature and public key evaluation is
		// important here since it can be distinguished by an
		// OP_CHECKMULTISIG NOT when the strict encoding flag is set.

		rawSig := sigInfo.signature
		if len(rawSig) == 0 {
			// Skip to the next pubkey if signature is empty.
			continue
		}

		// Split the signature into hash type and signature components.
		hashType := SigHashType(rawSig[len(rawSig)-1])
		signature := rawSig[:len(rawSig)-1]

		// Only parse and check the signature encoding once.
		var parsedSig *btcec.Signature
		if !sigInfo.parsed {
			if err := vm.checkHashTypeEncoding(hashType); err != nil {
				return err
			}
			if err := vm.checkSignatureEncoding(signature); err != nil {
				return err
			}

			// Parse the signature.
			var err error
			if vm.hasFlag(ScriptVerifyStrictEncoding) ||
				vm.hasFlag(ScriptVerifyDERSignatures) {

				parsedSig, err = btcec.ParseDERSignature(signature,
					btcec.S256())
			} else {
				parsedSig, err = btcec.ParseSignature(signature,
					btcec.S256())
			}
			sigInfo.parsed = true
			if err != nil {
				continue
			}
			sigInfo.parsedSignature = parsedSig
		} else {
			// Skip to the next pubkey if the signature is invalid.
			if sigInfo.parsedSignature == nil {
				continue
			}

			// Use the already parsed signature.
			parsedSig = sigInfo.parsedSignature
		}

		if err := vm.checkPubKeyEncoding(pubKey); err != nil {
			return err
		}

		// Parse the pubkey.
		parsedPubKey, err := btcec.ParsePubKey(pubKey, btcec.S256())
		if err != nil {
			continue
		}

		hash := calcScriptHash(script, hashType, &vm.tx, vm.txIdx)

		if parsedSig.Verify(hash, parsedPubKey) {
			// PubKey verified, move on to the next signature.
			signatureIdx++
			numSignatures--
		}
	}

	vm.dstack.PushBool(success)
	return nil
}

func opcodeCheckMultiSigVerify(op *parsedOpcode, vm *Engine) error {
	err := opcodeCheckMultiSig(op, vm)
	if err == nil {
		err = opcodeVerify(op, vm)
	}
	return err
}