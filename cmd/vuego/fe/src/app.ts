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

interface RefCallMessage {
  id?: number;
  method: string;
  params: {
    seq: number;
  };
}

(function() {
  let dev = true;

  class Vuego {
    ws: WebSocket;
    root: any; // {}
    resolveAPI: any;
    lastRefID: number;
    contextType: any;
    readyPromise: Promise<any>;
    beforeReady: () => void;

    constructor(ws: WebSocket) {
      this.ws = ws;
      this.resolveAPI = null;
      this.lastRefID = 0;
      this.beforeReady = null;

      const ready = new Promise((resolve, reject) => {
        this.resolveAPI = resolve;
      });
      this.readyPromise = ready;
      this.root = {
        Vuego(): Promise<any> {
          return ready;
        }
      };
      this.attach();
      this.initContext();
    }

    getapi(): any {
      return this.root;
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
          if (this.beforeReady !== null) {
            this.beforeReady();
          }
          if (this.resolveAPI != null) {
            this.resolveAPI(this.root);
          }
          break;
        }
      }
    }

    attach() {
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
    }

    bind(name: string) {
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
            callbacks.set(seq, args[i]); // root[bindingName].callbacks[callbackSeq] = func value
            args[i] = {
              bindingName: bindingName,
              seq: seq
            };
          } else if (args[i] instanceof this.contextType) {
            const seq = ++this.lastRefID;
            // js: rewrite input Context().seq = seq
            args[i].seq = seq;
            // go: will create Context object from seq and put it in jsclient.refs
            args[i] = {
              seq: seq
            };
          }
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

    initContext() {
      let $this = this;

      // Context class
      function Context() {
        this.seq = -1; // this will be rewrite as refID
        this.cancel = () => {
          let msg: RefCallMessage = {
            method: "Vuego.refCall",
            params: {
              seq: this.seq
            }
          };
          $this.ws.send(JSON.stringify(msg));
        };
        this.getThis = () => {
          return $this;
        };
      }
      this.contextType = Context;

      const TODO = new Context();
      const Backgroud = new Context();

      // context package
      this.root.context = {
        withCancel() {
          let ctx = new Context();
          return [ctx, ctx.cancel];
        },
        background() {
          return Backgroud;
        },
        todo() {
          return TODO;
        }
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

  function main() {
    let host = window.location.host;
    let ws = new WebSocket("ws://" + host + "/vuego");
    let vuego = new Vuego(ws);
    let api = vuego.getapi();

    let exportAPI = () => {
      let search = undefined;
      let name = getparam("name", search);
      let win: any = window;
      if (name === undefined || name === "window") Object.assign(win, api);
      else if (name) win[name] = api;
    };
    vuego.beforeReady = exportAPI;
    exportAPI();
  }
  main();
})();
