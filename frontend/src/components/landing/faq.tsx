"use client";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { faqs } from "./data";

export default function LandingFaq() {
  return (
    <section id="faq" className="bg-gray-50 py-24 dark:bg-gray-950">
      <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
        <div className="text-center">
          <span className="mb-4 inline-block rounded-full bg-brand-50 px-3 py-1 text-xs font-semibold text-brand-600 uppercase dark:bg-brand-500/15 dark:text-brand-400">
            FAQ
          </span>
          <h2 className="font-outfit text-3xl font-bold text-gray-900 sm:text-4xl lg:text-5xl dark:text-white">
            Questions,{" "}
            <span className="bg-gradient-to-r from-brand-500 to-brand-700 bg-clip-text text-transparent">
              Answered
            </span>
          </h2>
        </div>

        <div className="mt-10 rounded-2xl border border-gray-200 bg-white px-6 dark:border-gray-800 dark:bg-gray-900">
          <Accordion defaultValue={["q0"]}>
            {faqs.map((faq, i) => (
              <AccordionItem key={faq.question} value={`q${i}`}>
                <AccordionTrigger className="text-start text-base">
                  {faq.question}
                </AccordionTrigger>
                <AccordionContent className="text-sm leading-relaxed text-gray-600 dark:text-gray-400">
                  {faq.answer}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
      </div>
    </section>
  );
}
