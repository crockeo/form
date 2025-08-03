POST http://localhost:8000/api/v1/form
{
  "prompt": "hello world"
}
HTTP 200
[Captures]
form_id: jsonpath "$.id"

# ---
POST http://localhost:8000/api/v1/form/{{ form_id }}/response
{
  "text": "hello, other world"
}
HTTP 200

# ---

GET http://localhost:8000/api/v1/form/{{ form_id }}
HTTP 200
