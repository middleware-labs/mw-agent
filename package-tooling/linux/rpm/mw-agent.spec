Name: %{package_name}
Version: %{release_version}
Release: %{release_number}
Summary: Middleware Agent(%{package_name}) Service
License: GPL
Group: Development/Tools
Source0: %{package_name}-%{release_version}-%{arch}.tar.gz
Provides: %{package_name}
Obsoletes: %{package_name} <= %{release_version}

%description
Middleware Agent(%{package_name}) service enables you to monitor your infrastructure and applications.

%global __strip /bin/true

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/opt/%{package_name}/bin
mkdir -p %{buildroot}/etc/%{package_name}
mkdir -p %{buildroot}/lib/systemd/system
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name} %{buildroot}/opt/%{package_name}/bin/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/agent-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/otel-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/postinstall.sh %{buildroot}/opt/%{package_name}/.postinstall.sh
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name}.service %{buildroot}/lib/systemd/system/%{package_name}.service

%files
/opt/%{package_name}/bin/%{package_name}
/etc/%{package_name}/agent-config.yaml.sample
/etc/%{package_name}/otel-config.yaml.sample
/opt/%{package_name}/.postinstall.sh
/lib/systemd/system/%{package_name}.service

%post
chmod u+x /opt/%{package_name}/.postinstall.sh
/opt/%{package_name}/.postinstall.sh

%preun
if [ $1 -gt 0 ]; then
    echo "Upgrade in progress, skipping pre-uninstallation steps."
else
    systemctl stop %{package_name}
    systemctl disable %{package_name}
fi

%postun
if [ $1 -gt 0 ]; then
    echo "Upgrade in progress, skipping post-uninstallation steps."
else
    rm -f /etc/%{package_name}/agent-config.yaml
    rm -f /etc/%{package_name}/otel-config.yaml
    rmdir  /etc/%{package_name}
    rmdir /opt/%{package_name}/bin
    rmdir /opt/%{package_name}
fi

