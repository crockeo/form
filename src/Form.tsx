import { useParams } from "@solidjs/router";
import { For, Match, Switch, createResource, createSignal } from "solid-js";

interface Form {
  id: string;
  prompt: string;
  responses: FormResponse[];
}

interface FormResponse {
  id: string;
  text: string;
}

export default function Form() {
  const params = useParams();
  const [formDetails, { mutate }] = createResource(fetchForm);

  return (
    <div class="flex flex-col items-stretch min-h-screen mx-auto p-8 space-y-4 w-full sm:w-[40rem]">
      <Switch>
        <Match when={formDetails.loading}>
          <div>Loading...</div>
        </Match>

        <Match when={formDetails.error}>
          <div>Loading...</div>
        </Match>

        <Match when={formDetails()}>
          <FormBody form={formDetails()} respond={respond} />
        </Match>
      </Switch>
    </div>
  );

  async function fetchForm(): Promise<Form> {
    const res = await fetch(`/api/v1/form/${params.id}`);
    if (res.status !== 200) {
      // TODO: error handling
      return;
    }
    return await res.json();
  }

  async function respond(response: string): Promise<void> {
    const res = await fetch(`/api/v1/form/${params.id}/response`, {
      body: JSON.stringify({ text: response }),
      headers: { "Content-Type": "application/json" },
      method: "POST",
    });
    if (res.status !== 200) {
      // TODO: error handling
      return;
    }
    const form: Form = await res.json();
    mutate(form);
  }

  // async function handleSubmit(evt: SubmitEvent) {
  // 	evt.preventDefault();
  // 	const res = await fetch("/api/v1/form", {
  // 		body: JSON.stringify({ prompt: prompt() }),
  // 		headers: { "Content-Type": "application/json" },
  // 		method: "POST",
  // 	});
  // 	if (res.status != 200) {
  // 		// TODO: use solid toast and then set up a toast here.
  // 		return;
  // 	}
  // 	const body = await res.json();
  // 	console.log(body);
  // }
}

function FormBody(props: {
  form: Form;
  respond: (response: string) => Promise<void>;
}) {
  const [response, setResponse] = createSignal("");
  const disabled = () => response() === "";

  return (
    <div class="flex flex-col items-stretch justify-center space-y-4">
      <div class="font-bold text-4xl text-center">{props.form.prompt}</div>
      <form class="flex flex-row space-x-2" onsubmit={handleSubmit}>
        <input
          class="border grow p-2 rounded-lg"
          oninput={(e) => setResponse(e.target.value)}
          placeholder="Response"
          type="text"
          value={response()}
        />
        <button
          class="border p-2 rounded-lg transition"
          classList={{
            "opacity-25": disabled(),
            "hover:bg-black/5": !disabled(),
            "active:bg-black/10": !disabled(),
          }}
          disabled={disabled()}
          type="submit"
        >
          Respond
        </button>
      </form>

      <div>
        <div class="font-bold 2-xl">Responses so far:</div>

        <ul class="list-disc list-inside">
          <For each={props.form.responses}>
            {(response) => <li>{response.text}</li>}
          </For>
        </ul>
      </div>
    </div>
  );

  function handleSubmit(evt: SubmitEvent) {
    evt.preventDefault();
    try {
      props.respond(response());
    } catch {
      return;
    }
    setResponse("");
  }
}
