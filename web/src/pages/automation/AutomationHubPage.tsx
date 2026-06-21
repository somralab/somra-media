import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

export default function AutomationHubPage(): ReactNode {
  const { t } = useTranslation('automation');

  const sections = [
    {
      to: '/settings/automation/indexers',
      title: t('hub.indexers.title'),
      description: t('hub.indexers.description'),
    },
    {
      to: '/settings/automation/download-clients',
      title: t('hub.downloadClients.title'),
      description: t('hub.downloadClients.description'),
    },
    {
      to: '/settings/automation/quality-profiles',
      title: t('hub.qualityProfiles.title'),
      description: t('hub.qualityProfiles.description'),
    },
    {
      to: '/automation/downloads',
      title: t('hub.downloads.title'),
      description: t('hub.downloads.description'),
    },
    {
      to: '/settings/automation/monitors',
      title: t('hub.monitors.title'),
      description: t('hub.monitors.description'),
    },
  ];

  return (
    <section className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold">{t('hub.title')}</h1>
        <p className="text-sm text-muted">{t('hub.subtitle')}</p>
      </header>
      <div className="grid gap-4">
        {sections.map((section) => (
          <Card key={section.to}>
            <CardHeader>
              <CardTitle>{section.title}</CardTitle>
              <CardDescription>{section.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <Link to={section.to} className="text-sm text-primary hover:underline">
                {section.title}
              </Link>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
