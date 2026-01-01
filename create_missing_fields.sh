#!/bin/bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMTAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAxIiwibmFtZSI6ImFkbWluQHRlc3QuY29tIiwiZW1haWwiOiJhZG1pbkB0ZXN0LmNvbSIsInByb2ZpbGVfaWQiOiJzeXN0ZW1fYWRtaW4ifSwiZXhwIjoxNzY3MjUxNDU0LCJpYXQiOjE3NjcxNjUwNTQsImp0aSI6IjkxMTRkMTUwLTczZDQtNDk0NS04Mjg1LTFjNjhlMmM2MzhlNSJ9.p7n6qJ5ldjnaMZWLRDUoEM8jntt3GGNeFIWmnOh01Rw"
API_URL="http://localhost:3001/api/metadata/objects"

echo "Creating Account Fields..."

# Account: Phone
curl -s -X POST "$API_URL/account/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Phone", "api_name": "phone", "type": "Text"}'

# Account: Website
curl -s -X POST "$API_URL/account/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Website", "api_name": "website", "type": "Url"}'

# Account: Type
curl -s -X POST "$API_URL/account/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Type", "api_name": "type", "type": "Picklist", "options": ["Customer", "Partner", "Prospect", "Vendor"]}'

# Account: Billing Address
curl -s -X POST "$API_URL/account/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Billing Address", "api_name": "billing_address", "type": "LongText"}'

echo "Creating Contact Fields..."

# Contact: First Name
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "First Name", "api_name": "first_name", "type": "Text", "required": true}'

# Contact: Last Name
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Last Name", "api_name": "last_name", "type": "Text", "required": true}'

# Contact: Email
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Email", "api_name": "email", "type": "Email"}'

# Contact: Phone
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Phone", "api_name": "phone", "type": "Text"}'

# Contact: Title
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Title", "api_name": "title", "type": "Text"}'

# Contact: Account Lookup
curl -s -X POST "$API_URL/contact/fields"   -H "Authorization: Bearer $TOKEN"   -H "Content-Type: application/json"   -d '{"label": "Account", "api_name": "account_id", "type": "Lookup", "reference_to": ["account"]}'

echo "Done."
