import gleam/crypto
import gleam/int
import gleam/string
import gleam/time/timestamp

// A lot of this code is inspired by youid:
// https://github.com/lpil/youid/blob/main/src/youid/uuid.gleam

pub opaque type Uuid {
  Uuid(value: BitArray)
}

pub fn v7() -> Uuid {
  let #(sec, ns) =
    timestamp.system_time()
    |> timestamp.to_unix_seconds_and_nanoseconds()

  v7_from_millisec(sec * 1000 + ns / 1_000_000)
}

pub fn v7_from_millisec(timestamp: Int) -> Uuid {
  let assert <<a:size(12), b:size(62), _:size(6)>> =
    crypto.strong_random_bytes(10)
  let value = <<timestamp:48, 7:4, a:12, 2:2, b:62>>
  Uuid(value: value)
}

pub fn v7_string() -> String {
  v7() |> to_string
}

pub fn to_string(uuid: Uuid) -> String {
  let separator = "-"
  to_string_help(uuid.value, 0, "", separator)
}

fn to_string_help(
  ints: BitArray,
  position: Int,
  acc: String,
  separator: String,
) -> String {
  case position {
    8 | 13 | 18 | 23 ->
      to_string_help(ints, position + 1, acc <> separator, separator)
    _ ->
      case ints {
        <<i:size(4), rest:bits>> -> {
          let string = int.to_base16(i) |> string.lowercase
          to_string_help(rest, position + 1, acc <> string, separator)
        }
        _ -> acc
      }
  }
}
