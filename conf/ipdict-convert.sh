#!/bin/sh

#cat ipdict.txt  | awk -F, '{if(NF!=12&&NF!=14){print $0}}'
cat ipdict.txt  | awk -F, '{if(NF==14){print $1","$2","$3","$4","$5","$6","$8","$9","$10","$11","$12","$14}else{print $0}}' > a.txt
