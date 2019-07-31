# github-test

DESCRUOTION
github-test is a simple program written using go and utilizing the go-getgithub library
API set. The program will add a license to all repositories a github organization that
the user has permission to do so.

INSTRUCTIONS
Here are the assumptions that are made for running the program successfully.
- Go must be set up properly when building the executable (the go-gitbub libraries must be
installed in the proper directory)
- The LICENSE (must be named LICENSE) file that will be added to the repositories must exist in the same directory
as the executable github-test
- You will be prompted to enter a valid user name and password to authenticate to github
- You will be prompted to enter an email address
- You will prompted to enter a valid organization that the user has membership in

If the above steps are correctly followed, then the LICENSE file will be added to a branch
and a push request will be created for the repository.

IMPROVEMENTS
- Changes need to be made to fetch the username and email address from the user programmatically
- Select from a list of licenses and fetch from proper location
- Better log messages
