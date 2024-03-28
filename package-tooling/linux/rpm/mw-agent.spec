Name: %{package_name}
Version: %{release_version}
BuildArch: x86_64 aarch64
Release: 1
Summary: Middleware Agent(%{package_name}) Service
License: GPL
Group: Development/Tools
Source0: %{package_name}-%{release_version}-%{arch}.tar.gz

%description
Middleware Agent(%{package_name}) service enables you to monitor your infrastructure and applications.

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/opt/%{package_name}/bin
mkdir -p %{buildroot}/etc/%{package_name}
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name} %{buildroot}/opt/%{package_name}/bin/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/agent-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/otel-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/postinstall.sh %{buildroot}/opt/%{package_name}/.postinstall.sh

%files
/opt/%{package_name}/bin/%{package_name}
/etc/%{package_name}/agent-config.yaml.sample
/etc/%{package_name}/otel-config.yaml.sample
/opt/%{package_name}/.postinstall.sh

%post
chmod u+x /opt/%{package_name}/.postinstall.sh
/opt/%{package_name}/.postinstall.sh
