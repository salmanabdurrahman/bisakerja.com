# Response Examples

Dokumen ini merapikan contoh endpoint, request, response, dan field reference yang sudah ada tanpa menghilangkan kontennya.

## `/customer/create`

Create a new customer associated with the user (account owner). This endpoint is used to create a customer that can be used for billing purposes. The created customer will be associated with the user (account owner) and can be used for billing and subscription management.

### cURL Request

```bash
curl --request POST 'https://api.mayar.id/hl/v1/customer/create' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here' \
--data-raw '{
"name": "Raihan Nur Azmii",
"email": "mraihanna19@gmail.com",
"mobile": "089912345678"
}
'
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "success",
  "data": {
    "name": "Raihan Nur Azmii",
    "email": "mraihanna19@gmail.com",
    "mobile": "089912345678",
    "userId": "348e083d-315a-4e5c-96b1-5a2a98c48413",
    "customerId": "b0356d4c-516a-403e-abfe-b144da7068b4"
  }
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Request Body

- `name` (`string`, required): Customer Name.
- `email` (`string`, required): Customer email.
- `mobile` (`string`, required): Customer telephone number.

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.
- `data` (`array of object`): The main data returned (detailed customer/user information).

##### `data` Structure (Array Of Object)

- `name` (`string`, required): Customer name.
- `email` (`string`, required): Customer email address.
- `mobile` (`string`, required): Customer phone number.
- `userId` (`string<uuid>`, required): Unique ID for the user (the account owner).
- `response` (`string<uuid>`, required): Unique ID for the customer associated with the user.

## `/invoice/create`

Create a new invoice for the customer. This endpoint is used to create an invoice that can be sent to the customer for payment. The invoice will be associated with the customer and can include details such as the amount, description, and due date.

### cURL Request

```bash
curl --request POST 'https://api.mayar.id/hl/v1/invoice/create' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here' \
--data-raw '{
"name": "andre jago",
"email": "alikusnadide@gmail.com",
"mobile": "085797522261",
"redirectUrl": "https://kelaskami.com/nexst23",
"description": "testing dulu pak",
"expiredAt": "2026-04-19T16:43:23.000Z",
"items":[{
"quantity": 3,
"rate": 11000,
"description": "1e 1 sayam jago"
}],
"extraData": {
"noCustomer": "827hiueqy271hj",
"idProd": "contoh aja"
}
}'
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "success",
  "data": {
    "id": "df65d192-8396-4f9a-b4e5-8244648c07c5",
    "transactionId": "ca87fd13-8742-4d48-af33-7de1a417bc34",
    "link": "https://korban-motivator.mayar.shop/invoices/ycfyxbj2h3",
    "expiredAt": 1776617003000,
    "extraData": {
      "noCustomer": "827hiueqy271hj",
      "idProd": "contoh aja"
    }
  }
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Request Body

- `name` (`string`, required): Full name of the customer.
- `email` (`string`, required): Email address of the customer.
- `mobile` (`string`, required): Mobile phone number of the customer.
- `redirectUrl` (`string`, required): URL where the customer will be redirected after completing the payment.
- `description` (`string`, required): Description or notes related to the invoice.
- `expiredAt` (`string`, required): Invoice expiration time in ISO 8601 format (UTC).
- `items` (`array`, required): List of invoice items.
- `items[].quantity` (`number`, required): Quantity of the item.
- `items[].rate` (`number`, required): Price per item.
- `items[].description` (`string`, required): Description of the item.
- `extraData` (`object`, required): Additional custom data attached to the invoice.
- `extraData.noCustomer` (`string`, required): Custom customer reference number.
- `extraData.idProd` (`string`, required): Custom product identifier associated with the invoice.

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.
- `data` (`array of object`): The main data returned (invoice details).

##### `data` Structure (Array Of Object)

- `data.id` (`string`, required): Unique identifier of the invoice record.
- `data.transactionId` (`string`, required): Unique identifier of the associated transaction.
- `data.link` (`string`, required): Invoice URL that can be accessed by the customer.
- `data.expiredAt` (`number`, required): Timestamp (in milliseconds) indicating when the invoice will expire.
- `data.extraData` (`object`, required): Additional custom data attached to the invoice.
- `data.extraData.noCustomer` (`string`, required): Custom customer reference number.
- `data.extraData.idProd` (`string`, required): Custom product identifier associated with the invoice.

## `/coupon/validate`

Validate a coupon code. This endpoint is used to check if a given coupon code is valid and can be applied to an invoice. The response will indicate whether the coupon code is valid, and if valid, it will return the discount details associated with the coupon code.

### cURL Request

```bash
curl --request GET 'https://api.mayar.id/hl/v1/coupon/validate' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here' \
--data '{
"paymentLinkId": "4d26ea37-d093-4b92-8f5f-0faec64d65b0",
"tickets": [],
"couponCode": "NFRBFUK",
"finalAmount": 0,
"customerEmail": ""
}'
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "Success",
  "data": {
    "valid": true,
    "coupon": {
      "id": "08d233e9-91bc-4d9b-afc2-1dbbe38a1186",
      "code": "143KYCN",
      "discountType": "percentage",
      "discountValue": 10,
      "minimumPurchase": null,
      "eligibleCustomerType": "all"
    }
  }
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Request Body

- `paymentLinkId` (`string`, required): Unique identifier of the payment link.
- `tickets` (`array`): List of selected tickets for the checkout. Leave empty if not applicable.
- `couponCode` (`string`, required): Coupon code applied to the checkout. Leave empty if not used.
- `finalAmount` (`integer`, required): Final amount to be paid after discounts or adjustments.
- `customerEmail` (`string`): Email address of the customer performing the checkout.

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.
- `data` (`array of object`): Main response data.

##### `data` Structure (Array Of Object)

- `valid` (`boolean`, required): Indicates whether the coupon is valid or not.
- `coupon` (`object`, required): Object containing coupon details.
- `coupon.id` (`string`, required): Unique identifier of the coupon.
- `coupon.code` (`string`, required): Coupon code applied to the checkout.
- `coupon.discountType` (`string`, required): Type of discount applied (e.g., percentage, fixed).
- `coupon.discountValue` (`number`, required): Discount value based on the discount type.
- `coupon.minimumPurchase` (`number | null`, required): Minimum purchase amount required to use the coupon. Returns null if not set.
- `coupon.eligibleCustomerType` (`string`, required): Type of customers eligible to use the coupon (e.g., all, new, existing).

## `webhook/history?page=1&pageSize=10`

Get webhook event history. This endpoint is used to retrieve a list of past webhook events that have been triggered by the system. The response will include details about each webhook event, such as the event type, payload, timestamp, and delivery status. This can be useful for debugging and monitoring webhook integrations.

### cURL Request

```bash
curl --request GET 'https://api.mayar.id/hl/v1/webhook/history?page=1&pageSize=10' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here'
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "success",
  "hasMore": true,
  "pageCount": 77,
  "pageSize": 10,
  "page": 1,
  "data": [
    {
      "id": "7d567063-ad7f-48d5-9e84-0e41938783a5",
      "createdAt": 1745919352229,
      "nextDelivery": null,
      "paymentLinkId": "688b6a9f-2893-4b8a-a637-a008d91d0cfc",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"43b2f0ce-03f2-4f59-a341-299ea3ef19b6\",\"transactionId\":\"43b2f0ce-03f2-4f59-a341-299ea3ef19b6\",\"status\":\"SUCCESS\",\"transactionStatus\":\"paid\",\"createdAt\":\"2025-04-29T09:35:43.635Z\",\"updatedAt\":\"2025-04-29T09:35:43.635Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"ladu\",\"customerEmail\":\"ladusing@gmail.com\",\"customerMobile\":\"0888775540242\",\"amount\":0,\"isAdminFeeBorneByCustomer\":null,\"isChannelFeeBorneByCustomer\":null,\"productId\":\"688b6a9f-2893-4b8a-a637-a008d91d0cfc\",\"productName\":\"E-Book JOURNAL DESIGN OF AUTOMATIC\",\"productType\":\"ebook\",\"pixelFbp\":\"fb.1.1716364188358.1111830624\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745919352229,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "43b2f0ce-03f2-4f59-a341-299ea3ef19b6",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "7c98876e-bed4-44de-bd37-b144a4559476",
      "createdAt": 1745918541638,
      "nextDelivery": null,
      "paymentLinkId": "f361e6f5-5e92-4ba8-a244-b9b7823003fc",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"8e537afc-4f2f-4e1c-b91b-f75af34bf71a\",\"transactionId\":\"8e537afc-4f2f-4e1c-b91b-f75af34bf71a\",\"status\":\"SUCCESS\",\"transactionStatus\":\"paid\",\"createdAt\":\"2025-04-29T09:22:11.239Z\",\"updatedAt\":\"2025-04-29T09:22:11.239Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"ladu\",\"customerEmail\":\"ladusing@gmail.com\",\"customerMobile\":\"0888775540242\",\"amount\":0,\"isAdminFeeBorneByCustomer\":null,\"isChannelFeeBorneByCustomer\":null,\"productId\":\"f361e6f5-5e92-4ba8-a244-b9b7823003fc\",\"productName\":\"Test 04\",\"productType\":\"digital_product\",\"pixelFbp\":\"fb.1.1716364188358.1111830624\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745918541638,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "8e537afc-4f2f-4e1c-b91b-f75af34bf71a",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "6e93eae5-e190-4989-8e0f-1b4870be9bce",
      "createdAt": 1745569215626,
      "nextDelivery": null,
      "paymentLinkId": "d071116a-4359-4262-90dc-a7bd137c1baa",
      "payload": "{\"event\":\"payment.reminder\",\"data\":{\"id\":\"089fcdc2-5292-4e09-8b29-cfb7bdde3be6\",\"transactionId\":\"089fcdc2-5292-4e09-8b29-cfb7bdde3be6\",\"status\":\"SUCCESS\",\"transactionStatus\":\"created\",\"createdAt\":\"2025-04-25T07:58:27.539Z\",\"updatedAt\":\"2025-04-25T08:00:21.672Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Harus Buat auto testing\",\"customerEmail\":\"mayartesting05@gmail.com\",\"customerMobile\":\"081320547877\",\"amount\":107135,\"isAdminFeeBorneByCustomer\":true,\"isChannelFeeBorneByCustomer\":true,\"productId\":\"d071116a-4359-4262-90dc-a7bd137c1baa\",\"productName\":\"test UIUX Produk Digital\",\"productType\":\"digital_product\",\"pixelFbp\":\"fb.1.1737105110463.389414050868219356\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":\"qris\",\"paymentUrl\":\"https://andiak.myr.id/select-channel/089fcdc2-5292-4e09-8b29-cfb7bdde3be6?webhook=true\"}}",
      "status": "SUCCESS",
      "type": "payment.reminder",
      "updatedAt": 1745569215626,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "089fcdc2-5292-4e09-8b29-cfb7bdde3be6",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "af46a267-8896-446d-ac7e-90c212e756db",
      "createdAt": 1745386516347,
      "nextDelivery": null,
      "paymentLinkId": "0dd01bc9-ca99-4c0f-94df-fcfb78b835de",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"2c53458c-51ee-4fcc-8f6c-03b4ee948054\",\"transactionId\":\"2c53458c-51ee-4fcc-8f6c-03b4ee948054\",\"status\":\"SUCCESS\",\"transactionStatus\":\"paid\",\"createdAt\":\"2025-04-23T05:35:05.951Z\",\"updatedAt\":\"2025-04-23T05:35:05.951Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Harus Buat auto testing\",\"customerEmail\":\"mayartesting05@gmail.com\",\"customerMobile\":\"081320547877\",\"amount\":0,\"isAdminFeeBorneByCustomer\":null,\"isChannelFeeBorneByCustomer\":null,\"productId\":\"0dd01bc9-ca99-4c0f-94df-fcfb78b835de\",\"productName\":\" Algoritma Pemrograman Dasar\",\"productType\":\"course\",\"pixelFbp\":\"fb.1.1737105110463.389414050868219356\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745386516348,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "2c53458c-51ee-4fcc-8f6c-03b4ee948054",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "7514ea9b-61e7-4d3f-af42-7f7fc08cfaa6",
      "createdAt": 1745386309689,
      "nextDelivery": null,
      "paymentLinkId": "f361e6f5-5e92-4ba8-a244-b9b7823003fc",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"e801de8c-ab24-4ffa-ac14-99349495c325\",\"transactionId\":\"e801de8c-ab24-4ffa-ac14-99349495c325\",\"status\":\"SUCCESS\",\"transactionStatus\":\"paid\",\"createdAt\":\"2025-04-23T05:31:39.805Z\",\"updatedAt\":\"2025-04-23T05:31:39.805Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Harus Buat auto testing\",\"customerEmail\":\"mayartesting05@gmail.com\",\"customerMobile\":\"081320547877\",\"amount\":0,\"isAdminFeeBorneByCustomer\":null,\"isChannelFeeBorneByCustomer\":null,\"productId\":\"f361e6f5-5e92-4ba8-a244-b9b7823003fc\",\"productName\":\"Test 04\",\"productType\":\"digital_product\",\"pixelFbp\":\"fb.1.1737105110463.389414050868219356\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745386309689,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "e801de8c-ab24-4ffa-ac14-99349495c325",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "544df424-7584-478b-8ec9-37e4507d3296",
      "createdAt": 1745386189675,
      "nextDelivery": null,
      "paymentLinkId": "4f1af572-9a12-489f-b6bb-a067ab1402d5",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"f90f4d32-bdf7-4d11-8253-6f7ce19e4a78\",\"transactionId\":\"f90f4d32-bdf7-4d11-8253-6f7ce19e4a78\",\"status\":\"SUCCESS\",\"transactionStatus\":\"created\",\"createdAt\":\"2025-04-23T05:28:45.172Z\",\"updatedAt\":\"2025-04-23T05:28:46.785Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Harus Buat auto testing\",\"customerEmail\":\"mayartesting05@gmail.com\",\"customerMobile\":\"081320547877\",\"amount\":1045,\"isAdminFeeBorneByCustomer\":true,\"isChannelFeeBorneByCustomer\":true,\"productId\":\"4f1af572-9a12-489f-b6bb-a067ab1402d5\",\"productName\":\"test UIUX Link Pembayaran\",\"productType\":\"generic_link\",\"pixelFbp\":\"fb.1.1737105110463.389414050868219356\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":\"QRIS\",\"nettAmount\":1000,\"custom_field\":[{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"text\",\"isRequired\":false,\"key\":\"6df52b26-b018-4855-ab3d-8fdfb46861b8\",\"type\":\"string\",\"value\":\"tests\"},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"multipleText\",\"isRequired\":false,\"key\":\"80912ec7-e83d-46c6-afab-55ad39e88114\",\"type\":\"string\",\"value\":\"belajara\"},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"checkbox\",\"isRequired\":false,\"key\":\"55695415-e3bb-4127-92e0-29c7a756812d\",\"type\":\"lists\",\"value\":[{\"name\":\"Harus Buat auto testing\",\"value\":\"55695415-e3bb-4127-92e0-29c7a756812d\",\"selected\":true}]},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"textNumber\",\"isRequired\":false,\"key\":\"ddcd3c04-928e-4dd2-b011-5d5919a622e3\",\"type\":\"number\",\"value\":\"1232\"},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"date\",\"isRequired\":false,\"key\":\"fffe0331-8813-4e63-9802-8a3fab49368d\",\"type\":\"datetime\",\"value\":\"2025-04-25T17:00:00.000Z\"},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"select\",\"isRequired\":false,\"key\":\"0922d967-5635-43ce-855e-cb66e375047e\",\"type\":\"lists\",\"value\":[{\"name\":\"ddfdf\",\"value\":\"ddfdf\",\"key\":\"9f66d951-b773-4992-bccd-03586ffb865d\",\"selected\":false},{\"name\":\"perempuan\",\"value\":\"perempuan\",\"key\":\"51f180ec-70be-4588-83ba-f3faca6deea1\",\"selected\":false},{\"name\":\"laki\",\"value\":\"laki\",\"key\":\"c25f5360-c3e8-47d7-940c-d2f60d788225\",\"selected\":true}],\"selectedValue\":{\"name\":\"laki\",\"value\":\"laki\",\"key\":\"c25f5360-c3e8-47d7-940c-d2f60d788225\",\"selected\":true}},{\"name\":\"Harus Buat auto testing\",\"description\":\"Baskomg\",\"fieldType\":\"url\",\"isRequired\":false,\"key\":\"a598d6d6-e39c-4f30-8755-62141567c848\",\"type\":\"url\",\"value\":\"https://andiak.myr.id/pl/test-uiux-55576\"}]}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745386189676,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "f90f4d32-bdf7-4d11-8253-6f7ce19e4a78",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "897920cd-4dd8-4d40-af96-0c8c541fac43",
      "createdAt": 1745378929327,
      "nextDelivery": null,
      "paymentLinkId": "670ee967-8d09-40d4-95b2-7983f6ce5544",
      "payload": "{\"event\":\"payment.received\",\"data\":{\"id\":\"43bf4898-c081-4e5d-a6da-f5c78ec46486\",\"transactionId\":\"43bf4898-c081-4e5d-a6da-f5c78ec46486\",\"status\":\"SUCCESS\",\"transactionStatus\":\"paid\",\"createdAt\":\"2025-04-23T03:28:39.894Z\",\"updatedAt\":\"2025-04-23T03:28:39.895Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantName\":\"Andika\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Aldoghandika\",\"customerEmail\":\"aldobaik18@gmail.com\",\"customerMobile\":\"0877828291999\",\"amount\":0,\"isAdminFeeBorneByCustomer\":null,\"isChannelFeeBorneByCustomer\":null,\"productId\":\"670ee967-8d09-40d4-95b2-7983f6ce5544\",\"productName\":\"test UIUX Membership\",\"productType\":\"membership\",\"pixelFbp\":\"fb.1.1737105110463.389414050868219356\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null,\"membershipTier\":{\"id\":\"3a57c51e-48d8-4051-9651-6eff0a5eae24\",\"createdAt\":1738212685739,\"paymentAtStart\":true,\"paymentLinkId\":\"670ee967-8d09-40d4-95b2-7983f6ce5544\",\"updatedAt\":1738212685739,\"gracePeriodInDays\":0,\"isTrialAvailable\":true,\"finishMembershipAt\":null,\"name\":\"test UIUX 3\",\"notes\":\"Lorem ipsum dolor sit amet consectetur adipiscing, elit posuere euismod nascetur purus morbi, nam commodo fames sociis non. Litora aliquam imperdiet vulputate iaculis quisque eleifend nam lectus eget facilisi maecenas ullamcorper, sed malesuada nisi est sollicitudin magnis aptent blandit lacus donec. Sed placerat volutpat natoque ullamcorper cubilia senectus, elementum nisi tristique venenatis himenaeos metus felis, faucibus scelerisque cras arcu pretium. Nisl tincidunt nullam imperdiet scelerisque condimentum lectus inceptos, risus suspendisse purus sociosqu hendrerit consequat arcu neque, erat urna donec nunc dapibus maecenas. Nostra id pellentesque bibendum porttitor cubilia etiam luctus platea nec, porta magnis tortor fringilla venenatis conubia enim nisi, dictumst habitasse eros integer dictum praesent accumsan euismod. \",\"limit\":null},\"membershipCustomer\":{\"id\":\"3d22701d-0811-406e-b289-f7c92ce04a89\",\"isInTrial\":true,\"expiredAt\":1747970919291,\"status\":\"active\",\"isLifetimePeriod\":null,\"nextPayment\":1747970919291,\"monthlyPaymentPeriod\":null,\"membershipTierId\":\"3a57c51e-48d8-4051-9651-6eff0a5eae24\",\"isAlreadyUsedTrial\":null}}}",
      "status": "SUCCESS",
      "type": "payment.received",
      "updatedAt": 1745378929327,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "43bf4898-c081-4e5d-a6da-f5c78ec46486",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "8153c243-c5bf-4145-941e-9aef43011677",
      "createdAt": 1745377230152,
      "nextDelivery": null,
      "paymentLinkId": "670ee967-8d09-40d4-95b2-7983f6ce5544",
      "payload": "{\"event\":\"payment.reminder\",\"data\":{\"id\":\"d43eaa25-01b7-461e-9340-ba95b453cb84\",\"transactionId\":\"d43eaa25-01b7-461e-9340-ba95b453cb84\",\"status\":\"SUCCESS\",\"transactionStatus\":\"created\",\"createdAt\":\"2025-04-23T02:31:30.865Z\",\"updatedAt\":\"2025-04-23T02:35:31.483Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"ali kusnadi\",\"customerEmail\":\"kugutsu.hiruko@gmail.com\",\"customerMobile\":\"6285797522261\",\"amount\":33000,\"isAdminFeeBorneByCustomer\":false,\"isChannelFeeBorneByCustomer\":false,\"productId\":\"670ee967-8d09-40d4-95b2-7983f6ce5544\",\"productName\":\"test UIUX Membership\",\"productType\":\"membership\",\"pixelFbp\":\"fb.1.1745375116069.518591344665233937\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":null,\"paymentUrl\":\"https://andiak.myr.id/plt/d43eaa25-01b7-461e-9340-ba95b453cb84?webhook=true\"}}",
      "status": "SUCCESS",
      "type": "payment.reminder",
      "updatedAt": 1745377230152,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "d43eaa25-01b7-461e-9340-ba95b453cb84",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "539e936e-5fa2-4710-b8fd-50863c188921",
      "createdAt": 1745377228533,
      "nextDelivery": null,
      "paymentLinkId": "670ee967-8d09-40d4-95b2-7983f6ce5544",
      "payload": "{\"event\":\"payment.reminder\",\"data\":{\"id\":\"c238d027-fe8c-4777-9241-48c635b66597\",\"transactionId\":\"c238d027-fe8c-4777-9241-48c635b66597\",\"status\":\"SUCCESS\",\"transactionStatus\":\"created\",\"createdAt\":\"2025-04-23T02:26:35.509Z\",\"updatedAt\":\"2025-04-23T02:30:24.027Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"ali kusnadi\",\"customerEmail\":\"kugutsu.hiruko@gmail.com\",\"customerMobile\":\"6285797522261\",\"amount\":13782,\"isAdminFeeBorneByCustomer\":true,\"isChannelFeeBorneByCustomer\":true,\"productId\":\"670ee967-8d09-40d4-95b2-7983f6ce5544\",\"productName\":\"test UIUX Membership\",\"productType\":\"membership\",\"pixelFbp\":\"fb.1.1745375116069.518591344665233937\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":\"qris\",\"paymentUrl\":\"https://andiak.myr.id/select-channel/c238d027-fe8c-4777-9241-48c635b66597?webhook=true\"}}",
      "status": "SUCCESS",
      "type": "payment.reminder",
      "updatedAt": 1745377228533,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "c238d027-fe8c-4777-9241-48c635b66597",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    },
    {
      "id": "b13dbe97-1dd6-4383-ab92-400f9cdd3b74",
      "createdAt": 1745377224450,
      "nextDelivery": null,
      "paymentLinkId": "805ceb62-54c9-4671-9b2a-4d13c643b78e",
      "payload": "{\"event\":\"payment.reminder\",\"data\":{\"id\":\"8850d886-39a7-42f7-927c-16e8fb346d16\",\"transactionId\":\"8850d886-39a7-42f7-927c-16e8fb346d16\",\"status\":\"SUCCESS\",\"transactionStatus\":\"created\",\"createdAt\":\"2025-04-23T02:30:29.693Z\",\"updatedAt\":\"2025-04-23T02:35:31.235Z\",\"merchantId\":\"8afa8430-9ccd-4d33-9c1b-e0999a319994\",\"merchantEmail\":\"mayartesting05@gmail.com\",\"customerName\":\"Firzha Ardhia Ramadhan\",\"customerEmail\":\"zhaardhiatesting@gmail.com\",\"customerMobile\":\"08551498798\",\"amount\":1073,\"isAdminFeeBorneByCustomer\":true,\"isChannelFeeBorneByCustomer\":true,\"productId\":\"805ceb62-54c9-4671-9b2a-4d13c643b78e\",\"productName\":\"test UIUX Audiobook\",\"productType\":\"audiobook\",\"pixelFbp\":\"fb.1.1721824347392.952216048745555286\",\"pixelFbc\":null,\"qty\":1,\"couponUsed\":null,\"paymentMethod\":\"qris\",\"paymentUrl\":\"https://andiak.myr.id/select-channel/8850d886-39a7-42f7-927c-16e8fb346d16?webhook=true\"}}",
      "status": "SUCCESS",
      "type": "payment.reminder",
      "updatedAt": 1745377224450,
      "userId": "8afa8430-9ccd-4d33-9c1b-e0999a319994",
      "paymentLinkTransactionId": "8850d886-39a7-42f7-927c-16e8fb346d16",
      "urlDestination": "https://eoze6x1sc8k58q4.m.pipedream.net",
      "responsePayload": null,
      "source": "AUTOMATIC"
    }
  ]
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Query Parameters

- `page` (`integer`, default: `1`, required): Determine which page.
- `pageSize` (`integer`, default: `10`, required): How much data is displayed on the page.

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.
- `hasMore` (`boolean`, required): Information on whether there is additional data to be retrieved (true/false).
- `pageCount` (`integer`, required): Total number of data pages.
- `pageSize` (`integer`, required): Number of items per page.
- `page` (`integer`, required): The page currently displayed.
- `data` (`array of object`): List of products displayed.

##### `data` Structure (Array Of Object)

- `id` (`string<uuid>`, required): Webhook delivery record ID.
- `createdAt` (`integer (timestamp)`, required): Creation time of the webhook delivery log.
- `nextDelivery` (`string | null`, required): Timestamp of the next delivery attempt (if any).
- `paymentLinkId` (`string<uuid>`, required): Associated payment link ID.
- `payload` (`string (JSON string)`, required): Original webhook payload sent (contains event + transaction data).
- `status` (`string`, required): Delivery status (example: `"SUCCESS"`).
- `type` (`string`, required): Webhook event type (example: `"payment.received"`, `"payment.reminder"`).
- `updatedAt` (`integer (timestamp)`, required): Last update timestamp.
- `userId` (`string`, required): User ID who owns this webhook.
- `paymentLinkTransactionId` (`string<uuid>`, required): Associated transaction ID.
- `urlDestination` (`string (URL)`, required): Destination endpoint where the webhook is delivered.
- `responsePayload` (`string | null`, required): Response returned from the destination endpoint.
- `source` (`string`, required): Indicates whether the webhook was sent automatically or manually (example: `"AUTOMATIC"`).

## `/webhook/register`

Register a webhook URL. This endpoint allows you to register a webhook URL where the system will send notifications about specific events, such as payment updates or invoice status changes. By providing a valid URL, you can ensure that your application receives real-time updates and can respond accordingly to events related to your invoices and transactions.

### cURL Request

```bash
curl --request GET 'https://api.mayar.id/hl/v1/webhook/register' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here' \
--data '{
"urlHook": "https://example.mayar.com"
}'
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "success"
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Request Body

- `urlHook` (`string`, required): URL for webhook callback.

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.

## `/invoice/{id}`

Get invoice details by ID. This endpoint is used to retrieve the details of a specific invoice using its unique identifier (ID). The response will include information about the invoice such as the customer details, invoice items, amount, status, and other relevant data. This can be useful for displaying invoice information to customers or for internal record-keeping and management purposes.

### cURL Request

```bash
curl --request GET 'https://api.mayar.id/hl/v1/invoice/f774034d-d9cc-43a0-97d8-a2520c127f03' \
--header 'Authorization: Bearer Paste-Your-API-Key-Here' \
--data ''
```

### Response Example

```json
{
  "statusCode": 200,
  "messages": "success",
  "data": {
    "id": "f774034d-d9cc-43a0-97d8-a2520c127f03",
    "amount": 110000,
    "status": "unpaid",
    "link": "ibzfrf4880",
    "expiredAt": 1764582069401,
    "transactions": [
      {
        "id": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa"
      }
    ],
    "customerId": "ae57ce73-89a2-46a7-84d7-93a616ef220e",
    "customer": {
      "id": "ae57ce73-89a2-46a7-84d7-93a616ef220e",
      "email": "azumiikecee@gmail.com",
      "mobile": "08996136751",
      "name": "Azumii"
    },
    "transactionId": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa",
    "paymentUrl": "https://andiak.myr.id/invoices/ibzfrf4880",
    "paymentLinkId": "f774034d-d9cc-43a0-97d8-a2520c127f03"
  }
}
```

### Authorization

- `Authorization` (`string`, required)
  - Example: `Authorization | Bearer Paste-Your-API-Key-Here`

### Path Parameters

- `id` (`string<uuid>`, required): A unique identifier obtained when creating an invoice. If accessed through the dashboard, it can be found at the URL on the invoice details page.
  - Example: `f774034d-d9cc-43a0-97d8-a2520c127f03`

### Response

#### Successful Response

##### Main Structure (Root)

- `statusCode` (`integer`, required): Status code form API.
- `messages` (`string`, required): Status message that describes the status code.
- `data` (`array of object`): List of invoices displayed.

##### `data` Structure (Array Of Object)

- `id` (`string<uuid>`, required): Unique invoice ID.
- `amount` (`integer`, required): Invoice price or nominal value.
- `status` (`string`, required): Product status.
- `link` (`string`, required): Product slug/short URL.
- `type` (`string`, required): Product type.
- `expiredAt` (`integer`, required): Invoice expiration time in timestamp form (epoch millis).
- `transactions` (`array of object`): List of invoice related transactions.
- `customerId` (`string<uuid>`, required): Unique customer ID.
- `customer` (`array of object`): Details of the customer who paid the invoice.
- `transactionId` (`string<uuid>`, required): The primary transaction ID for this invoice.
- `paymentUrl` (`string`, required): URL for invoice payment page.
- `paymentLinkId` (`string<uuid>`, required): Unique ID of payment link (usually the same as invoice ID).

##### `transactions` Structure (Array Of Object)

- `id` (`string <uuid>`, required): Unique transaction ID.

##### `customer` Structure (Array Of Object)

- `id` (`string <uuid>`, required): Unique customer ID.
- `email` (`string`, required): Customer's email address.
- `mobile` (`string`, required): Customer's mobile phone number.
- `name` (`string`, required): Customer's name.
