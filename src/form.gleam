import gleam/erlang/process
import gleam/http
import gleam/io
import mist
import sqlight
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
      )
      ",
      conn,
    )

  let secret_key_base = wisp.random_string(64)
  let assert Ok(_) =
    wisp_mist.handler(handle_request(conn), secret_key_base)
    |> mist.new
    |> mist.port(8000)
    |> mist.start

  io.println("Listening on :8000")
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
  todo
}

fn handle_get_room(
  conn: sqlight.Connection,
  req: wisp.Request,
  id: string,
) -> wisp.Response {
  todo
}

fn handle_post_response(
  conn: sqlight.Connection,
  req: wisp.Request,
  id: string,
) -> wisp.Response {
  todo
}
