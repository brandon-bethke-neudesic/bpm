package main;


import (
    "fmt"
)

type HelpCommand struct {
}


func (help *HelpCommand) Execute() (error){
    fmt.Println("");
    fmt.Println("Usage: bpm <command>")
    fmt.Println("");
    fmt.Println("where <command> is one of: ")
    fmt.Println("");
    fmt.Println("    install")
    fmt.Println("");
    fmt.Println("        bpm install [url] [commit] ( [--remote=myremote] | [--root=diskpath] )")
    fmt.Println("");
    fmt.Println("        Examples:");
    fmt.Println("");
    fmt.Println("        # install all the dependencies in the bpm.json file using the remote origin as the root path if necessary")
    fmt.Println("        bpm install");
    fmt.Println("");
    fmt.Println("        # install and save the mortar dependency using the remote origin as root path and the latest commit.")
    fmt.Println("        bpm install ../mortar.git");
    fmt.Println("");
    fmt.Println("        # install and save the mortar dependency using the remote origin as root path and the specified commit.")
    fmt.Println("        bpm install ../mortar.git ad2c7c47362fc682079307cbb1db7ef944997364")
    fmt.Println("");
    fmt.Println("        # install and save the mortar dependency using the remote brandon as root path and the latest commit.")
    fmt.Println("        bpm install ../mortar.git --remote=brandon");
    fmt.Println("");
    fmt.Println("        # install and save the mortar dependency using the specified url and the latest commit.")
    fmt.Println("        bpm install https://neudesic.timu.com/projects/timu/code/master/mortar.git");
    fmt.Println("");
    fmt.Println("        # install all the dependencies in the bpm.json file and use the specified folder as the root path if necessary")
    fmt.Println("        # ignores commit information")
    fmt.Println("        bpm install --root=../js");
    fmt.Println("");
    fmt.Println("    update")
    fmt.Println("")
    fmt.Println("        bpm update [dependencyName] ( [--remote=myremote] | [--root=diskpath] ) [--recursive]");
    fmt.Println("");
    fmt.Println("        Examples:")
    fmt.Println("");
    fmt.Println("        # update the all the dependencies in the bpm.json.")
    fmt.Println("        bpm update");
    fmt.Println("");
    fmt.Println("        # update the existing mortar dependency to the latest commit. Use the remote origin as the root path if necessary.")
    fmt.Println("        bpm update mortar");
    fmt.Println("");
    fmt.Println("        # update the existing mortar dependency to the latest commit. Use the remote brandon as the root path if necessary.")
    fmt.Println("        bpm update mortar --remote=brandon");
    fmt.Println("");
    fmt.Println("        # update the existing mortar dependency to the latest commit in the specified path")
    fmt.Println("        bpm update mortar --root=../js");
    fmt.Println("");
    fmt.Println("        # update all dependencies recursively. Only works with the --root option")
    fmt.Println("        bpm update mortar --root=../js --recursive");
    fmt.Println("")
    fmt.Println("    init")
    fmt.Println("")
    fmt.Println("        bpm init <modulename>");
    fmt.Println("");
    fmt.Println("        Examples:")
    fmt.Println("");
    fmt.Println("        # create the default bpm.json file with the name my-module")
    fmt.Println("        bpm init my-module");
    fmt.Println("")
    fmt.Println("    clean")
    fmt.Println("")
    fmt.Println("        bpm clean");
    fmt.Println("");
    fmt.Println("        Examples:")
    fmt.Println("");
    fmt.Println("        # delete the bpm_modules folder")
    fmt.Println("        bpm clean");
    fmt.Println("")
    fmt.Println("    ls")
    fmt.Println("")
    fmt.Println("        bpm ls");
    fmt.Println("");
    fmt.Println("        Examples:")
    fmt.Println("");
    fmt.Println("        # list the installed dependencies")
    fmt.Println("        bpm ls");
    fmt.Println("");
    fmt.Println("    help")
    fmt.Println("")
    fmt.Println("        bpm help");
    fmt.Println("");
    return nil;
}
