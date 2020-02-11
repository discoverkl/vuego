import { getapi, Base } from "vue2go";

// -------------------------------------------------
// api and mock
// -------------------------------------------------
interface API extends Base {
  mock: boolean;
  SIGINT: number;
  SIGKILL: number;

  name(): string;
  write(s: string): void;
  listen(writer: (string, number) => void): void;
  kill(sig: number): void;
  pwd(): Promise<string>;
  load(path: string): string;
  save(path: string, content: string);
}

let writer;
let api = {
  mock: true,
  write(s) {
    if (writer) {
      writer(s, 1);
      writer(s, 2);
    }
  },
  listen(w) {
    writer = w;
  },
  kill(sig) {
    api.write(`[SEND SIGNAL]: ${sig}\n`);
  },
  name() {
    return "bash";
  },
  pwd() {
    return new Promise((resolve, reject) => {
      let hold = this.pwd as any;
      const id = hold.id || 1;
      hold.id = id + 1;
      resolve("/home/user" + id);
    });
  },
  load(path) {
    if (path === "") return "";
    return `text content of path: ${path}`;
  },
  save(path, content) {}
} as API;

try {
  api = getapi() as API;
} catch (ex) {
  // console.error(ex);
}

// consts
api.SIGINT = 2;
api.SIGKILL = 9;

(window as any).api = api;

export default api;
