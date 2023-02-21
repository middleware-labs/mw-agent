Name: mw-go-agent-host-aws
Version: 0.0.1
Release: 1
Summary: My Package Summary
License: GPL
Group: Development/Tools
Source0: mw-go-agent-host-aws_0.0.1-1_all.tar.gz


%description
My package description

%prep
%setup -q

%build

%install
pwd
ls -l
tree mw-go-agent-host-aws-0.0.1
cp -rfa ~/rpmbuild/SOURCES/mw-go-agent-host-aws_0.0.1-1_all/mw-go-agent-host-aws-0.0.1/* %{buildroot}/usr/bin/

%files
/usr/bin/*
