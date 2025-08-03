import gleam/bool
import gleam/dynamic/decode
import gleam/erlang/process
import gleam/http
import gleam/io
import gleam/json
import gleam/list
import mist
import sqlight
import uuid
import wisp
import wisp/wisp_mist

pub fn main() -> Nil {
  wisp.configure_logger()

  use conn <- sqlight.with_connection(":memory:")
  let assert Ok(_) =
    sqlight.exec(
      "
      CREATE TABLE forms (
        id TEXT PRIMARY KEY,
        prompt TEXT
      );

      CREATE TABLE responses (
        id TEXT PRIMARY KEY,
        form_id TEXT NOT NULL,
        response TEXT NOT NULL,

        FOREIGN KEY (form_id) REFERENCES forms(id)
      );
      ",
      conn,
    )

  let secret_key_base = wisp.random_string(64)
  let assert Ok(_) =
    wisp_mist.handler(handle_request(conn), secret_key_base)
    |> mist.new
    |> mist.port(8000)
    |> mist.start

  process.sleep_forever()
}

fn middleware(
  req: wisp.Request,
  next: fn(wisp.Request) -> wisp.Response,
) -> wisp.Response {
  use <- wisp.log_request(req)
  use <- wisp.rescue_crashes
  next(req)
}

fn handle_request(conn: sqlight.Connection) -> fn(wisp.Request) -> wisp.Response {
  fn(req: wisp.Request) -> wisp.Response {
    use req <- middleware(req)
    case req.method, wisp.path_segments(req) {
      http.Post, ["api", "v1", "room"] -> handle_create_room(conn, req)
      http.Get, ["api", "v1", "room", id] -> handle_get_room(conn, req, id)
      http.Post, ["api", "v1", "room", id, "response"] ->
        handle_post_response(conn, req, id)

      _, _ -> wisp.not_found()
    }
  }
}

fn handle_create_room(
  conn: sqlight.Connection,
  req: wisp.Request,
) -> wisp.Response {
  let decoder = {
    use prompt <- decode.field("prompt", decode.string)
    decode.success(prompt)
  }
  use prompt <- require_decoded_json(req, decoder)

  let form_id = uuid.v7_string()
  use _ <- internal_server_error(sqlight.query(
    "INSERT INTO forms (id, prompt) VALUES (?, ?) RETURNING (id);",
    conn,
    [sqlight.text(form_id), sqlight.text(prompt)],
    decode.success(fn(a) { a }),
  ))

  wisp.json_response(
    json.to_string_tree(json.object([#("new_form_id", json.string(form_id))])),
    200,
  )
}

fn handle_get_room(
  conn: sqlight.Connection,
  _req: wisp.Request,
  id: String,
) -> wisp.Response {
  let decoder = {
    use form_prompt <- decode.field(0, decode.string)
    decode.success(form_prompt)
  }
  use form <- internal_server_error(sqlight.query(
    "SELECT prompt FROM forms WHERE id = ?;",
    conn,
    [sqlight.text(id)],
    decoder,
  ))
  let assert [form_prompt] = form
  use <- bool.guard(list.length(form) > 0, wisp.not_found())

  let decoder = {
    use response_id <- decode.field(0, decode.string)
    use response_text <- decode.field(1, decode.string)
    decode.success(#(response_id, response_text))
  }
  use responses <- internal_server_error(sqlight.query(
    "SELECT id, response FROM responses WHERE form_id = ?",
    conn,
    [sqlight.text(id)],
    decoder,
  ))

  wisp.json_response(
    json.to_string_tree(
      json.object([
        #("prompt", json.string(form_prompt)),
        #(
          "responses",
          json.preprocessed_array(
            list.map(responses, fn(response) {
              let #(response_id, response_text) = response
              json.object([
                #("id", json.string(response_id)),
                #("text", json.string(response_text)),
              ])
            }),
          ),
        ),
      ]),
    ),
    200,
  )
}

fn handle_post_response(
  conn: sqlight.Connection,
  req: wisp.Request,
  id: string,
) -> wisp.Response {
  todo
}

fn require_decoded_json(
  req: wisp.Request,
  decoder: decode.Decoder(t),
  next: fn(t) -> wisp.Response,
) -> wisp.Response {
  use json <- wisp.require_json(req)
  case decode.run(json, decoder) {
    Ok(t) -> next(t)
    Error(_) -> wisp.bad_request()
  }
}

fn internal_server_error(
  value: Result(t, u),
  next: fn(t) -> wisp.Response,
) -> wisp.Response {
  case value {
    Ok(t) -> next(t)
    Error(e) -> {
      echo e
      wisp.internal_server_error()
    }
  }
}
