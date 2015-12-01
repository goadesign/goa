// This module exports functions that give access to the cellar API hosted at cellar.goa.design.
// It uses the axios javascript library for making the actual HTTP requests.
define(['axios'] , function (axios) {
    return function (scheme, host, timeout) {
        scheme = scheme || 'https';
        host = host || 'cellar.goa.design';
        timeout = timeout || 20000;

        // Client is the object returned by this module.
        var client = axios;

        // URL prefix for all API requests.
        var urlPrefix = scheme + '://' + host;

        // Create new account
        // path is the request path, the format is "/cellar/accounts"
        // data contains the action payload (request body)
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.createAccount = function (path, data, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'post',
                data: data,
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // Record new bottle
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles"
        // data contains the action payload (request body)
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.createBottle = function (path, data, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'post',
                data: data,
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // deleteAccount calls the delete action of the account resource.
        // path is the request path, the format is "/cellar/accounts/:accountID"
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.deleteAccount = function (path, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'delete',
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // deleteBottle calls the delete action of the bottle resource.
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles/:bottleID"
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.deleteBottle = function (path, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'delete',
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // Retrieve account with given id
        // path is the request path, the format is "/cellar/accounts/:accountID"
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.showAccount = function (path, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'get',
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // Retrieve bottle with given id
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles/:bottleID"
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.showBottle = function (path, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'get',
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // Change account name
        // path is the request path, the format is "/cellar/accounts/:accountID"
        // data contains the action payload (request body)
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.updateAccount = function (path, data, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'put',
                data: data,
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // updateBottle calls the update action of the bottle resource.
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles/:bottleID"
        // data contains the action payload (request body)
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.updateBottle = function (path, data, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'patch',
                data: data,
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // List all bottles in account optionally filtering by year
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles"
        // years is used to build the request query string.
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.listBottle = function (path, years, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'get',
        params: {
                years: years
                    },
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }

        // rateBottle calls the rate action of the bottle resource.
        // path is the request path, the format is "/cellar/accounts/:accountID/bottles/:bottleID/actions/rate"
        // data contains the action payload (request body)
        // config is an optional object to be merged into the config built by the function prior to making the request.
        // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
        // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
        client.rateBottle = function (path, data, config) {
            cfg = {
                timeout: timeout,
                url: urlPrefix + path,
                method: 'put',
                data: data,
                responseType: 'json'
            };
            if (config) {
                cfg = utils.merge(cfg, config);
            }
            return axios(cfg);
        }
        return client;
    };
});
