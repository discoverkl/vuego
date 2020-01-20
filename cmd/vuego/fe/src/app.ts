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
  function attach(ws: WebSocket) {
    ws.onmessage = e => {
      let msg = JSON.parse(e.data);
      console.log("receive: ", JSON.stringify(msg, null, "  "));
      let method = msg.method;
      let params;
      switch (method) {
        case "Vuego.call":
          params = msg.params;
          switch (params.name) {
            case "eval":
              let ret, err;
              try {
                ret = eval(params.args[0]);
              } catch (ex) {
                err = ex.toString() || "unknown error";
              }
              // let retMsg: Message = {
              //   id: msg.id,
              //   method: "Vuego.ret",
              //   params: {
              //     result: ret,
              //     error: err
              //   }
              // };
              // ws.send(JSON.stringify(retMsg));
              reply(msg.id, ret, err);
              break;
          }
          break;
        case "Vuego.bind":
          params = msg.params;
          bind(params.name);
          break;
        case "Vuego.ready":
          // TODO: raise ready event only once
          break;
      }
    };

    ws.onopen = e => {
      // ws.send(JSON.stringify({ method: "1" }));
    };
    ws.onerror = e => {
      console.log("ws error:", e);
    };
    ws.onclose = e => {
      console.log("ws close:", e);
    };

    function reply(id: number, ret?: any, err?: string) {
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
      ws.send(JSON.stringify(msg));
    }
  }

  function getRoot(): any {
    if (window.vuego === undefined) {
      window.vuego = {};
    }
    return window.vuego;
  }

  function bind(name: string) {
    // const refBindingName = "%s";
    let root = getRoot();
    const bindingName = name;
    root[bindingName] = async (...args) => {
      const me = root[bindingName];

      // for (let i = 0; i < args.length; i++) {
      //   // support javascript functions as arguments
      //   if (typeof args[i] == "function") {
      //     let functions = me["functions"];
      //     if (!functions) {
      //       functions = new Map();
      //       me["functions"] = functions;
      //     }
      //     const seq = (functions["lastSeq"] || 0) + 1;
      //     functions["lastSeq"] = seq;
      //     functions.set(seq, args[i]);
      //     args[i] = {
      //       bindingName: bindingName,
      //       seq: seq
      //     };
      //   }
      //   // else if (args[i] instanceof context.Context) {
      //   //   const ref = root[refBindingName];
      //   //   let objs = ref["objs"];
      //   //   if (!objs) {
      //   //     objs = new Map();
      //   //     ref["objs"] = objs;
      //   //   }
      //   //   const seq = (objs["lastSeq"] || 0) + 1;
      //   //   objs["lastSeq"] = seq;
      //   //   args[i].seq = seq;
      //   //   args[i] = {
      //   //     seq: seq
      //   //   };
      //   // }
      // }

      let errors = me["errors"];
      let callbacks = me["callbacks"];
      if (!callbacks) {
        callbacks = new Map();
        me["callbacks"] = callbacks;
      }
      if (!errors) {
        errors = new Map();
        me["errors"] = errors;
      }
      const seq = (me["lastSeq"] || 0) + 1;
      me["lastSeq"] = seq;
      const promise = new Promise((resolve, reject) => {
        callbacks.set(seq, resolve);
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
      ws.send(JSON.stringify(callMsg));
      return promise;
    };
  }

  let host = window.document.location.host;
  let ws = new WebSocket(`ws://${host}/vuego`);
  attach(ws);
})();

async function main() {
  let sum = await vuego.add(2, 2);
  console.log(sum);
}
// main();
