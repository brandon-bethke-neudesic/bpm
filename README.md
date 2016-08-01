// me
{
    "name" : "login-client",
    "version" : "1.0.0",
    "dependencies" : {
        "mortar" : {
            "url" : "../mortar.git",
            "commit" : "XXXXXXX"
        },
        "null-query":{
            "url" :"../null-query.git",
            "commit" :"XXXXYY"
        }
    }
}

// mortar
{
    "name" : "mortar",
    "version" : "1.0.1",
    "dependencies" : {
        "null-query" : {
            "url" : "../null-query.git", // also support absolute urls
            "commit" : "XXXXXX"
        }
    }
}


// disk

nodule_modules
    null-query -> link to correct commit
    __bmp_tmp__
        null-query
            commit1
            commit2


bmp update module-name [--root=xxx] // get latest of module, put in node_modules, update json with commit, increment version number of json file, if root is specified look in file path root/module-name, update each sub dep in node_modules with --root logic

bpm install url [commit] [--remote=origin] // save bpm.json, increment version number 1.0.0 -> 1.0.1, place in node_modules

bpm install
