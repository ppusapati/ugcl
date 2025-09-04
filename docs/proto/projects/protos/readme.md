# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [projects/protos/dairy_site.proto](#projects_protos_dairy_site-proto)
    - [CreateDairySiteRequest](#projects-api-v2-dairy_site-CreateDairySiteRequest)
    - [CreateDairySiteResponse](#projects-api-v2-dairy_site-CreateDairySiteResponse)
    - [DairySite](#projects-api-v2-dairy_site-DairySite)
  
    - [DairySiteService](#projects-api-v2-dairy_site-DairySiteService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="projects_protos_dairy_site-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## projects/protos/dairy_site.proto



<a name="projects-api-v2-dairy_site-CreateDairySiteRequest"></a>

### CreateDairySiteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dairy_site | [DairySite](#projects-api-v2-dairy_site-DairySite) |  |  |






<a name="projects-api-v2-dairy_site-CreateDairySiteResponse"></a>

### CreateDairySiteResponse
Response with created DairySite


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dairy_site | [DairySite](#projects-api-v2-dairy_site-DairySite) |  |  |






<a name="projects-api-v2-dairy_site-DairySite"></a>

### DairySite
Custom JSONTime can be handled as google.protobuf.Timestamp


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name_of_site | [string](#string) |  |  |
| todays_work | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| latitude | [double](#double) |  |  |
| longitude | [double](#double) |  |  |
| submitted_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 

 

 


<a name="projects-api-v2-dairy_site-DairySiteService"></a>

### DairySiteService
Dairy service definition

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetAllDairySites | [.core.api.v2.query.GetQueryRequest](#core-api-v2-query-GetQueryRequest) | [.core.api.v2.query.GetQueryResponse](#core-api-v2-query-GetQueryResponse) |  |
| CreateDairySite | [CreateDairySiteRequest](#projects-api-v2-dairy_site-CreateDairySiteRequest) | [CreateDairySiteResponse](#projects-api-v2-dairy_site-CreateDairySiteResponse) |  |

 



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

