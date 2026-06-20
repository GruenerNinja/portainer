import { Form, Formik, FormikHelpers } from 'formik';
import { useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';

import { Widget } from '@@/Widget';

import { FormValues, formValuesToCreatePayload } from './type';
import { useCreateSourceMutation } from './useSourceCreateMutation';
import { WizardStep, useWizardContext } from './WizardContext';
import { WizardHeader } from './WizardHeader';
import { WizardFooter } from './WizardFooter';

const initialFormValues: FormValues = {
  name: '',
  type: 'git',
  git: {
    url: '',
    authentication: {
      authEnabled: true,
    },
    connectionOk: false,
  },
  vault: {
    address: '',
    namespace: '',
    kvVersion: 2,
    tlsSkipVerify: false,
    authentication: {
      method: 'token',
      token: '',
    },
    connectionOk: false,
  },
};

type Props = {
  steps: WizardStep[];
};

export function CreateForm({ steps }: Props) {
  const mutation = useCreateSourceMutation();
  const router = useRouter();
  const { currentStep, isLastStep, goToNextStep } = useWizardContext();

  return (
    <Formik
      initialValues={initialFormValues}
      onSubmit={handleSubmit}
      validationSchema={currentStep.validateStep}
    >
      <Form noValidate>
        <WizardHeader steps={steps} />
        <Widget>
          <currentStep.component />
        </Widget>
        <WizardFooter />
      </Form>
    </Formik>
  );

  function handleSubmit(
    formValues: FormValues,
    { setTouched, setSubmitting }: FormikHelpers<FormValues>
  ) {
    if (!isLastStep) {
      goToNextStep();
      setTouched({});
      setSubmitting(false);
      return;
    }

    mutation.mutate(formValuesToCreatePayload(formValues), {
      onSuccess: () => {
        notifySuccess('Success', 'Source successfully created');
        router.stateService.go('.^');
      },
      onSettled: () => setSubmitting(false),
    });
  }
}
