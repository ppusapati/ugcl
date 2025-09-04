# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [core/proto/query.proto](#core_proto_query-proto)
    - [FilterParam](#core-api-v2-query-FilterParam)
    - [GetQueryRequest](#core-api-v2-query-GetQueryRequest)
    - [GetQueryResponse](#core-api-v2-query-GetQueryResponse)
    - [QueryRecord](#core-api-v2-query-QueryRecord)
    - [QueryRecord.FieldsEntry](#core-api-v2-query-QueryRecord-FieldsEntry)
  
- [Scalar Value Types](#scalar-value-types)



<a name="core_proto_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## core/proto/query.proto



<a name="core-api-v2-query-FilterParam"></a>

### FilterParam



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| field | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="core-api-v2-query-GetQueryRequest"></a>

### GetQueryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page | [int32](#int32) |  |  |
| limit | [int32](#int32) |  |  |
| from_date | [string](#string) |  |  |
| to_date | [string](#string) |  |  |
| fields | [string](#string) | repeated |  |
| date_column | [string](#string) |  |  |
| filters | [FilterParam](#core-api-v2-query-FilterParam) | repeated |  |






<a name="core-api-v2-query-GetQueryResponse"></a>

### GetQueryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| total | [int64](#int64) |  |  |
| page | [int32](#int32) |  |  |
| limit | [int32](#int32) |  |  |
| data | [QueryRecord](#core-api-v2-query-QueryRecord) | repeated |  |






<a name="core-api-v2-query-QueryRecord"></a>

### QueryRecord



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [QueryRecord.FieldsEntry](#core-api-v2-query-QueryRecord-FieldsEntry) | repeated |  |






<a name="core-api-v2-query-QueryRecord-FieldsEntry"></a>

### QueryRecord.FieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

