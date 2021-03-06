package cli

import (
	"fmt"
)

func Help() {
	fmt.Println(`
Usage:

    clio COMMAND [FLAGS]

Commands:

    create      Creating application skeleton
    g           Generating scaffold, controller, router or view
    help        Output this message again

Generators:

    clio g [controller | view | router | scaffold] NAME

Example:

    clio create awesome-application

    This will generate a skeletal Clio application.
    Please, see the README in the newly created application
    get going. Also http://grsmv.github.io is a good place
    for getting more information about Clio. Good luck!
`)
}
