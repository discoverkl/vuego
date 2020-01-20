interface Message {
  id?: number;
  method: string;
  params: any;
}

interface CallMessage {
  id?: number;
  method: string;
  params: {
    name: string;
    seq: number;
    args: any[];
  };
}

(function() {
  let dev = true;

  class Vuego {
    ws: WebSocket;
    root: {};
    resolveAPI: any;

    constructor(ws: WebSocket) {
      this.ws = ws;
      this.root = {};
      this.resolveAPI = null;
    }

    replymessage(id: number, ret?: any, err?: string) {
      if (ret === undefined) ret = null;
      if (err === undefined) err = null;
      let msg: Message = {
        id: id,
        method: "Vuego.ret",
        params: {
          result: ret,
          error: err
        }
      };
      this.ws.send(JSON.stringify(msg));
    }

    onmessage(e: MessageEvent) {
      let ws = this.ws;
      let msg = JSON.parse(e.data);
      if (dev) console.log("receive: ", JSON.stringify(msg, null, "  "));
      let root = this.root;
      let method = msg.method;
      let params;
      switch (method) {
        case "Vuego.call": {
          params = msg.params;
          switch (params.name) {
            case "eval": {
              let ret, err;
              try {
                ret = eval(params.args[0]);
              } catch (ex) {
                err = ex.toString() || "unknown error";
              }
              this.replymessage(msg.id, ret, err);
              break;
            }
          }
          break;
        }
        case "Vuego.ret": {
          let { name, seq, result, error } = msg.params;
          if (error) {
            root[name]["errors"].get(seq)(error);
          } else {
            root[name]["results"].get(seq)(result);
          }
          root[name]["errors"].delete(seq);
          root[name]["results"].delete(seq);
          break;
        }
        case "Vuego.callback": {
          let { name, seq, args } = msg.params;
          let ret, err;
          try {
            ret = root[name]["callbacks"].get(seq)(...args);
          } catch (ex) {
            err = ex.toString() || "unknown error";
          }
          this.replymessage(msg.id, ret, err);
          break;
        }
        case "Vuego.closeCallback": {
          let { name, seq } = msg.params;
          root[name]["callbacks"].delete(seq);
          break;
        }
        case "Vuego.bind": {
          params = msg.params;
          this.bind(params.name);
          break;
        }
        case "Vuego.ready": {
          if (this.resolveAPI != null) {
            this.resolveAPI(this.root);
          }
          break;
        }
      }
    }

    async attach(): Promise<any> {
      let ws = this.ws;
      ws.onmessage = this.onmessage.bind(this);

      ws.onopen = e => {
        // ws.send(JSON.stringify({ method: "1" }));
      };

      ws.onerror = e => {
        console.log("ws error:", e);
      };

      ws.onclose = e => {
        console.log("ws close:", e);
      };

      // wait for ready
      const promise = new Promise((resolve, reject) => {
        this.resolveAPI = resolve;
      });
      return promise;
    }

    bind(name: string) {
      // const refBindingName = "%s";
      let root = this.root;
      const bindingName = name;
      root[bindingName] = async (...args) => {
        const me = root[bindingName];

        for (let i = 0; i < args.length; i++) {
          // support javascript functions as arguments
          if (typeof args[i] == "function") {
            let callbacks = me["callbacks"];
            if (!callbacks) {
              callbacks = new Map();
              me["callbacks"] = callbacks;
            }
            const seq = (callbacks["lastSeq"] || 0) + 1;
            callbacks["lastSeq"] = seq;
            callbacks.set(seq, args[i]); // root[bindingName].functions[callbackSeq] = func value
            args[i] = {
              bindingName: bindingName,
              seq: seq
            };
          }
          // else if (args[i] instanceof context.Context) {
          //   const ref = root[refBindingName];
          //   let objs = ref["objs"];
          //   if (!objs) {
          //     objs = new Map();
          //     ref["objs"] = objs;
          //   }
          //   const seq = (objs["lastSeq"] || 0) + 1;
          //   objs["lastSeq"] = seq;
          //   args[i].seq = seq;
          //   args[i] = {
          //     seq: seq
          //   };
          // }
        }

        // prepare (errors, results, lastSeq) on binding function
        let errors = me["errors"];
        let results = me["results"];
        if (!results) {
          results = new Map();
          me["results"] = results;
        }
        if (!errors) {
          errors = new Map();
          me["errors"] = errors;
        }
        const seq = (me["lastSeq"] || 0) + 1;
        me["lastSeq"] = seq;
        const promise = new Promise((resolve, reject) => {
          results.set(seq, resolve);
          errors.set(seq, reject);
        });

        // call go
        let callMsg: CallMessage = {
          method: "Vuego.call",
          params: {
            name: bindingName,
            seq,
            args
          }
        };
        // binding call phrase 1
        this.ws.send(JSON.stringify(callMsg));
        return promise;
      };
    }
  }

  function getparam(name: string, search?: string): string | undefined {
    search = search === undefined ? window.location.search : search;
    let pair = search
      .slice(1)
      .split("&")
      .map(one => one.split("="))
      .filter(one => one[0] == name)
      .slice(-1)[0];
    if (pair === undefined) return;
    return pair[1] || "";
  }

  async function main() {
    let host = window.location.host;
    let ws = new WebSocket("ws://" + host + "/vuego");
    let vuego = new Vuego(ws);
    let api = await vuego.attach();
    try {
    } catch (ex) {
      console.log("call error:", ex);
    }
    let search = undefined;
    let win: any = window;
    let name = getparam("name", search);
    if (name === undefined || name === "window") Object.assign(win, api);
    else if (name) win[name] = api;
  }
  main();
})();
