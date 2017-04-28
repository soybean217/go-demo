#!/bin/sh

#restart single and unique process
executeFileName="go-fm-register"
executeFileTimeStamp=`date +%s -r $executeFileName`
srcFileName="web.go"
srcFileTimeStamp=`date +%s -r $srcFileName`
logFileName="log.out"
psid=0

checkpid() {
  echo "checkpid $1"
  psid=0
  aaps=`ps -ef | grep $1 | grep -v grep`
  
  
  if [ -n "$aaps" ]; then
  psid=`echo $aaps | awk '{print $2}'`
  else
  psid=0
  fi
  
  echo "checkpided: $psid"
}

dt=`date +%Y%m%d-%H%M%S`;
if [[ $srcFileTimeStamp -gt executeFileTimeStamp ]]; then
  echo "begin backup ..."
  destFolder=/opt/backup/go-fm-register/$dt/;
  mkdir -p $destFolder;
  tar cvfz $destFolder/bin.tar.gz $executeFileName *.json;
  echo "end backup ."
  echo "begin build ."
  go build
  echo "end build ."
fi

checkpid $executeFileName ;
echo "kill $psid ..." ;
if [ $psid -ne 0 ];then
  kill $psid
fi
sleep 1s
mv logs/$logFileName logs/$logFileName.$dt
touch logs/$logFileName
nohup ./$executeFileName  > logs/$logFileName &
checkpid $executeFileName
echo "new pid: $psid ..."
