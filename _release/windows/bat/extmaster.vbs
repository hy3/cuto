' processName�Ɏw�肵���v���Z�X�������݂���ꍇ��1��Ԃ��B
' ���݂��Ȃ��ꍇ��0��Ԃ��B
processName = "master.exe"
If GetProcessId(processName) <> 0 Then
  WScript.Quit(1)
Else
  WScript.Quit(0)
End If

'-------------------------------------------------------------------------------
' �w�肳�ꂽ�v���Z�X����ID���擾����
Function GetProcessId(ProcessName)
    Dim Service,QfeSet,Qfe,r
    
    Set Service = WScript.CreateObject("WbemScripting.SWbemLocator").ConnectServer
    Set QfeSet = Service.ExecQuery("Select * From Win32_Process Where Caption='" & ProcessName & "'")
    
    r = 0
    For Each Qfe In QfeSet
        r = Qfe.ProcessId
        Exit For
    Next

    GetProcessId = r
End Function
