# ConcurrentRecursiveRenamer
Concurrently renames all files and directories in a specified directory according to the user specified rule

## Usage

To run the program simply use:

```bash
go run ConcurrentRecursiveRenamer.go directory target_string replacement_string
```

The program will recursively go through the specified directory and rename each file by replacing the target_string with the replacement_string.

E.g.:

Before usage:  
├── root_dir  
│&emsp;├── branch_1  
│&emsp;│&emsp;├── branch_1_doc.txt  
│&emsp;├── branch_2  
│&emsp;│&emsp;├── branch_2_2  
│&emsp;│&emsp;│&emsp;├── deep_branch.c  
│&emsp;├── document.txt  
│&emsp;├── images  


```bash
go run ConcurrentRecursiveRenamer.go root_dir branch child
```

After usage:

├── root_dir  
│&emsp;├── child_1  
│&emsp;│&emsp;├── child_1_doc.txt  
│&emsp;├── child_2  
│&emsp;│&emsp;├── child_2_2  
│&emsp;│&emsp;│&emsp;├── deep_child.c  
│&emsp;├── document.txt  
│&emsp;├── images  

## 

When the program is ran, the main thread spawns a go routine for each file in the specified directory.  
Each go routine will then calculate a new path for the current file and if it is a directory it will spawn a new go routine which will do the same thing recursively until all the files of the root directory have been explored.  
When a go routine reaches a file that isn't a directory it sends the new calculated path to the go routine that originally spawned it, resulting in all the new paths being sent to the main thread. Main then orders the paths by depth and renames then sequentially to not cause any naming conflicts.
