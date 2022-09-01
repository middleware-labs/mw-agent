const http = require('http');
var request = require('request');


if (process.env.MW_AUTOGENERATE_TRACING_DATA) {
    setTimeout(() => {}, 5000);

    setInterval(()=>{

        let random = Math.random();

        if (random < 0.9) {
            http.get('http://localhost:3002/api/tutorials');
        }

        if (random < 0.1) {
            http.get('http://localhost:3002/error');
        }

        if (random < 0.3) {
            http.get('http://localhost:3002/500-error');
        }

        if (random < 0.5) {
            http.get('http://localhost:3002/504-error');
        }

        if (random < 0.6) {
            http.get('http://localhost:3002/304-error');
        }

        if (random < 0.7) {
            request.post('http://localhost:3002/api/tutorials', {
                "title": "git3",
                "description": "test description3"
            }, (error, response, body) => {
                let body_new = JSON.parse(body);
                request.put(`http://localhost:3002/api/tutorials/${body_new.id}`, {
                    "title": "git4",
                    "description": "test description4"
                    }, (puterror, putresponse, putbody) => {

                        request.delete(`http://localhost:3002/api/tutorials/${body_new.id}`);
                });
            });
        }

    },3000);
}