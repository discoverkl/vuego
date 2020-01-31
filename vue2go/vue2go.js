/*!
 * vue2go.js v1.0.6
 * (c) 2019-2020 Leo Kong
 * Released under the MIT License.
 */
(function(global, factory) {
  typeof exports === "object" && typeof module !== "undefined"
    ? (module.exports = factory())
    : typeof define === "function" && define.amd
    ? define(factory)
    : ((global = global || self), (global.Vuego = factory()));
})(this, function() {
  "use strict";

  let apiCache = null;
  let apiPromiseCache = null;

  // Create a rpc client which can call the go server binding methods directly from javascript.
  //
  // interface options {
  //   async: boolean      // if true, return promise (default: false)
  //   publicPath: string  // prefix path of the vuego.js file serving from go server (default: '/')
  // }
  function getapi(options) {
    if (!options) options = {};
    const async = options.async === undefined ? false : options.async;
    const publicPath =
      options.publicPath === undefined ? "/" : options.publicPath;

    const path = `${publicPath}vuego.js?name=`;

    if (async) {
      if (apiPromiseCache === null) {
        apiPromiseCache = new Promise((reslove, reject) => {
          fetch(path)
            .then(reply => {
              reply
                .text()
                .then(script => {
                  if (!reply.ok) {
                    reject(
                      new Error(
                        `connect vuego server failed: ${reply.status} ${reply.statusText}`
                      )
                    );
                  }
                  try {
                    const api = eval(script);
                    reslove(api);
                  } catch (ex) {
                    reject(new Error(`import vuego failed: ${ex}`));
                  }
                })
                .catch(ex => reject(ex));
            })
            .catch(ex => reject(ex));
        });
      }
      return apiPromiseCache;
    } else {
      if (apiCache === null) {
        // [Deprecation] Synchronous XMLHttpRequest on the main thread is deprecated
        // because of its detrimental effects to the end user's experience.
        // For more help, check https://xhr.spec.whatwg.org/.
        var request = new XMLHttpRequest();
        request.open("GET", path, false);
        request.send(null);
        if (request.status < 200 || request.status > 299) {
          throw new Error(
            `connect vuego server failed: ${request.status} ${request.statusText}`
          );
        }

        try {
          apiCache = eval(request.responseText);
        } catch (ex) {
          throw new Error(`import vuego failed: ${ex}`);
        }
      }
      return apiCache;
    }
  }

  return {
    getapi
  };
});
